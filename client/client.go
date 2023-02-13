package client

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/tests"
	"github.com/jrcamenzuli/network-performance-tester-client/types"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const testResultsDirectory string = "test-results/"

func createLogFile(filename string, contents func(*csv.Writer)) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Could not create " + filename)
	}
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	contents(w)
}

func RunClient(config *types.Configuration) {

	logfilePrefix := strings.Replace(strings.Replace(time.Now().UTC().Format(time.RFC3339), ":", "", -1), "-", "", -1)
	config.Client.LogfilePostfix = "-" + config.Client.LogfilePostfix

	logConfigInfo(logfilePrefix, config)

	logDeviceInfo(logfilePrefix, config.Client.LogfilePostfix)

	if config.Client.Tests.IdleStateOfDevice.Enable {
		testIdleStateOfDevice(logfilePrefix, config.Client.LogfilePostfix)
	}

	if config.Client.Tests.IdleStateOfProcess.Enable {
		testIdleStateOfProcess(logfilePrefix, config.Client.LogfilePostfix, config.Client.PID)
	}

	if config.Client.Tests.HTTP_Throughput.Enable {
		fmt.Println("Starting HTTP Throughput Test")
		testHTTP_Throughput(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTP_Port,
			config.Client.PID,
			false)
	}

	if config.Client.Tests.HTTPS_Throughput.Enable {
		fmt.Println("Starting HTTPS Throughput Test")
		testHTTP_Throughput(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTPS_Port,
			config.Client.PID,
			true)
	}

	if config.Client.Tests.Ping.Enable {
		fmt.Println("Starting Ping Test")
		testPing(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerPingPort,
			config.Client.Tests.Ping.CountSamples)
	}

	if config.Client.Tests.Jitter.Enable {
		fmt.Println("Starting Jitter Test")
		testJitter(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerPingPort,
			config.Client.Tests.Jitter.CountDifferences)
	}

	if config.Client.Tests.HTTP_Burst.Enable {
		fmt.Println("Starting HTTP Burst Test")
		testHTTP_Burst(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTP_Port,
			10, // countTestsToRun
			func(i int) int { return (i + 1) * 10 }, config.Client.PID,
			false)
	}

	if config.Client.Tests.HTTPS_Burst.Enable {
		fmt.Println("Starting HTTPS Burst Test")
		testHTTP_Burst(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTPS_Port,
			10, // countTestsToRun
			func(i int) int { return (i + 1) * 10 }, config.Client.PID,
			true)
	}

	if config.Client.Tests.HTTP_Rate.Enable {
		fmt.Println("Starting HTTP Rate Test")
		testHTTP_Rate(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTP_Port,
			time.Second*1, // restDuration
			5,             // countTestsToRun
			func(i int) int { return (i + 1) * 10 },
			time.Second*time.Duration(config.Client.Tests.HTTP_Rate.Duration), // testDuration
			config.Client.PID,
			false)
	}
	if config.Client.Tests.HTTPS_Rate.Enable {
		fmt.Println("Starting HTTPS Rate Test")
		testHTTP_Rate(
			logfilePrefix,
			config.Client.LogfilePostfix,
			config.Client.ServerHost,
			config.Client.ServerTCP_HTTPS_Port,
			time.Second*1, // restDuration
			5,             // countTestsToRun
			func(i int) int { return (i + 1) * 10 },
			time.Second*time.Duration(config.Client.Tests.HTTPS_Rate.Duration), // testDuration
			config.Client.PID,
			true)
	}
}

func logConfigInfo(logfilePrefix string, config *types.Configuration) {
	postfix := config.Client.LogfilePostfix
	f, err := os.Create("./test-results/" + logfilePrefix + "-configInfo" + postfix + ".txt")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	out := util.PrettifyStruct(config)
	fmt.Printf("Config:\n%s\n\n", out)
	f.Write([]byte(out))
}

func logDeviceInfo(logfilePrefix string, logfilePostfix string) {
	cpus, err := cpu.Info() // the CPU description
	if err != nil || len(cpus) <= 0 {
		str := fmt.Sprintf("\nI could not retrieve device information to log.\n\n")
		fmt.Printf(util.WarningColor, str)
		return
	}
	cpu := cpus[0]
	mem, _ := mem.VirtualMemory() // the system memory description
	deviceInfo := model.DUT_Info{CPU_ModelName: cpu.ModelName, CPU_CoreCount: uint(cpu.Cores), CPU_BaseClockFrequency: uint(cpu.Mhz) * 1e6, RAM_Total: uint(mem.Total)}
	f, err := os.Create("./test-results/" + logfilePrefix + "-deviceInfo" + logfilePostfix + ".txt")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	out := util.PrettifyStruct(deviceInfo)
	fmt.Printf("Device Info:\n%s\n\n", out)
	f.Write([]byte(out))
}

// Idle state test of device
func testIdleStateOfDevice(logfilePrefix string, logfilePostfix string) {
	filename := testResultsDirectory + logfilePrefix + "-idleStateOfDevice" + logfilePostfix + ".csv"
	time.Sleep(1 * time.Second)
	cpuUsage := tests.IdleStateOfDevice()
	fmt.Printf("Idle state of device: CPU %.2f%%\n", cpuUsage*100.0)
	contents := func(w *csv.Writer) {
		w.Write([]string{"CPU (%)"})
		cpu := fmt.Sprintf("%.4f%%", cpuUsage)
		w.Write([]string{cpu})
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

// Idle state test of a process
func testIdleStateOfProcess(logfilePrefix string, logfilePostfix string, pid uint) {
	filename := testResultsDirectory + logfilePrefix + "-idleStateOfProcess" + logfilePostfix + ".csv"
	idleStateOfProcess := tests.IdleStateOfProcess(pid)
	fmt.Printf("Idle state of \"%d\" process: CPU %.2f%%, RAM %dMB\n", pid, idleStateOfProcess.Cpu*100.0, idleStateOfProcess.Ram/1e6)
	if idleStateOfProcess.Cpu == 0 && idleStateOfProcess.Ram == 0 {
		str := fmt.Sprintf("\nI could not monitor a process with PID \"%d\" because it could not be found.\n\n", pid)
		fmt.Printf(util.WarningColor, str)
	}
	if idleStateOfProcess == nil {
		return
	}
	contents := func(w *csv.Writer) {
		w.Write([]string{"CPU (%)", "RAM (MB)"})
		cpu := fmt.Sprintf("%.4f%%", idleStateOfProcess.Cpu)
		ram := fmt.Sprintf("%d", idleStateOfProcess.Ram/1e6)
		w.Write([]string{cpu, ram})
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

// HTTP Burst test barrage
func testHTTP_Burst(logfilePrefix string, logfilePostfix string, serverHost string, serverPort uint, countTestsToRun int, fn model.Fn, pid uint, isHttps bool) {
	serverProtocol := ""
	testNameForFile := ""
	if isHttps {
		serverProtocol = "https://"
		testNameForFile = "-httpsBurstTest"
	} else {
		serverProtocol = "http://"
		testNameForFile = "-httpBurstTest"
	}
	url := fmt.Sprintf("%s%s:%d/download/100000", serverProtocol, serverHost, serverPort)
	filename := testResultsDirectory + logfilePrefix + testNameForFile + logfilePostfix + ".csv"
	contents := func(w *csv.Writer) {
		w.Write([]string{"number of http requests in burst", "time to complete (ms)", "failure rate (%)", "average CPU (%)", "average RAM (MB)"}) // todo: log CPU and RAM too
		for i := 0; i < countTestsToRun; i++ {
			burstSize := fn(i)
			result := tests.HttpBurstTest(url, burstSize, pid, isHttps)
			failureRate := fmt.Sprintf("%.4f", result.FailureRate)
			var cpu, ram string
			if result.CpuAndRam.Ram != 0 {
				cpu = fmt.Sprintf("%.4f", result.CpuAndRam.Cpu)
				ram = fmt.Sprintf("%d", result.CpuAndRam.Ram/1e6)
			}
			if err := w.Write([]string{strconv.Itoa(burstSize), strconv.Itoa(int(result.Duration.Milliseconds())), failureRate, cpu, ram}); err != nil {
				log.Fatalln("error writing record to file", err)
			}
			w.Flush()
		}
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

// HTTP Rate test barrage
func testHTTP_Rate(logfilePrefix string, logfilePostfix string, serverHost string, serverPort uint, restDuration time.Duration, countTestsToRun int, fn model.Fn, testDuration time.Duration, pid uint, isHttps bool) {
	serverProtocol := ""
	testNameForFile := ""
	if isHttps {
		serverProtocol = "https://"
		testNameForFile = "-httpsRateTest"
	} else {
		serverProtocol = "http://"
		testNameForFile = "-httpRateTest"
	}
	url := fmt.Sprintf("%s%s:%d/download/1000", serverProtocol, serverHost, serverPort)
	filename := testResultsDirectory + logfilePrefix + testNameForFile + logfilePostfix + ".csv"
	contents := func(w *csv.Writer) {
		w.Write([]string{"requests per second", "test duration (ms)", "average CPU (%)", "average RAM (MB)", "failure rate (%)"})

		for i := 0; i < countTestsToRun; i++ {
			requestsPerSecond := fn(i)
			result := tests.HttpRateTest(url, testDuration, requestsPerSecond, pid, isHttps)
			var cpu, ram string
			if result.CpuAndRam.Ram != 0 {
				cpu = fmt.Sprintf("%.4f", result.CpuAndRam.Cpu)
				ram = fmt.Sprintf("%d", result.CpuAndRam.Ram/1e6)
			}
			failureRate := fmt.Sprintf("%.4f", result.FailureRate)
			w.Write([]string{strconv.Itoa(requestsPerSecond), strconv.Itoa(int(testDuration.Milliseconds())), cpu, ram, failureRate})
			w.Flush()
		}
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

func testHTTP_Throughput(logfilePrefix string, logfilePostfix string, serverHost string, serverPort uint, pid uint, isHttps bool) {
	serverProtocol := ""
	testNameForFile := ""
	if isHttps {
		serverProtocol = "https://"
		testNameForFile = "-httpThroughputTest"
	} else {
		serverProtocol = "http://"
		testNameForFile = "-httpsThroughputTest"
	}
	fmt.Printf("Half Duplex Throughput:\n")
	filename := testResultsDirectory + logfilePrefix + testNameForFile + logfilePostfix + ".csv"
	contents := func(w *csv.Writer) {
		w.Write([]string{"transfer mode (half/full duplex)", "bytes transferred (MB)", "duration (ms)", "transfer rate (MB/s)", "transfer rate (Mb/s)", "average CPU (%)", "average RAM (MB)"})

		uploadThroughputTestResult, _ := tests.UploadThroughputTest(serverProtocol, serverHost, serverPort, pid)
		Bps := float64(uploadThroughputTestResult.CountBytesTransferred) / (float64(uploadThroughputTestResult.DurationNanoseconds) / 1e9)
		bps := Bps * 8
		fmt.Printf("%s\t--------- %.0fMB @ %.0fMB/s (%.0fMb/s) ------------\n", uploadThroughputTestResult.Type, float64(uploadThroughputTestResult.CountBytesTransferred)/1e6, Bps/1e6, bps/1e6)

		transferMode := fmt.Sprintf("%s", uploadThroughputTestResult.Type)
		bytes_MB := fmt.Sprintf("%.0f", float64(uploadThroughputTestResult.CountBytesTransferred)/1e6)
		duration_ms := fmt.Sprintf("%.0f", (float64(uploadThroughputTestResult.DurationNanoseconds) / 1e6))
		rate_MBps := fmt.Sprintf("%.0f", Bps/1e6)
		rate_Mbps := fmt.Sprintf("%.0f", bps/1e6)
		var cpu, ram string
		if uploadThroughputTestResult.CpuAndRam.Ram != 0 {
			cpu = fmt.Sprintf("%.4f", uploadThroughputTestResult.CpuAndRam.Cpu)
			ram = fmt.Sprintf("%d", uploadThroughputTestResult.CpuAndRam.Ram/1e6)
		}
		w.Write([]string{transferMode, bytes_MB, duration_ms, rate_MBps, rate_Mbps, cpu, ram})
		w.Flush()

		downloadThroughputTestResult, _ := tests.DownloadThroughputTest(serverProtocol, serverHost, serverPort, pid)
		Bps = float64(downloadThroughputTestResult.CountBytesTransferred) / (float64(downloadThroughputTestResult.DurationNanoseconds) / 1e9)
		bps = Bps * 8
		fmt.Printf("%s\t--------- %.0fMB @ %.0fMB/s (%.0fMb/s) ------------\n", downloadThroughputTestResult.Type, float64(downloadThroughputTestResult.CountBytesTransferred)/1e6, Bps/1e6, bps/1e6)

		transferMode = fmt.Sprintf("%s", downloadThroughputTestResult.Type)
		bytes_MB = fmt.Sprintf("%.0f", float64(downloadThroughputTestResult.CountBytesTransferred)/1e6)
		duration_ms = fmt.Sprintf("%.0f", (float64(downloadThroughputTestResult.DurationNanoseconds) / 1e6))
		rate_MBps = fmt.Sprintf("%.0f", Bps/1e6)
		rate_Mbps = fmt.Sprintf("%.0f", bps/1e6)
		if downloadThroughputTestResult.CpuAndRam.Ram != 0 {
			cpu = fmt.Sprintf("%.4f", uploadThroughputTestResult.CpuAndRam.Cpu)
			ram = fmt.Sprintf("%d", uploadThroughputTestResult.CpuAndRam.Ram/1e6)
		}
		w.Write([]string{transferMode, bytes_MB, duration_ms, rate_MBps, rate_Mbps, cpu, ram})
		w.Flush()

		fmt.Printf("\n")

		fmt.Printf("Full Duplex Throughput:\n")
		results := make(chan model.ThroughputTest, 2)
		errors := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := tests.DownloadThroughputTest(serverProtocol, serverHost, serverPort, pid)
			result.Type = model.RX_FullDuplex
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := tests.UploadThroughputTest(serverProtocol, serverHost, serverPort, pid)
			result.Type = model.TX_FullDuplex
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()

		go func() {
			wg.Wait()
			close(results)
			close(errors)
		}()
		for err := range errors {
			fmt.Printf("%s\n", err.Error())
			return

		}

		for throughputTestResult := range results {
			Bps := float64(throughputTestResult.CountBytesTransferred) / (float64(throughputTestResult.DurationNanoseconds) / 1e9)
			bps := Bps * 8
			fmt.Printf("%s\t--------- %.0fMB @ %.0fMB/s (%.0fMb/s) ------------\n", throughputTestResult.Type, float64(throughputTestResult.CountBytesTransferred)/1e6, Bps/1e6, bps/1e6)

			transferMode = fmt.Sprintf("%s", throughputTestResult.Type)
			bytes_MB = fmt.Sprintf("%.0f", float64(throughputTestResult.CountBytesTransferred)/1e6)
			duration_ms = fmt.Sprintf("%.0f", (float64(throughputTestResult.DurationNanoseconds) / 1e6))
			rate_MBps = fmt.Sprintf("%.0f", Bps/1e6)
			rate_Mbps = fmt.Sprintf("%.0f", bps/1e6)
			if uploadThroughputTestResult.CpuAndRam.Ram != 0 {
				cpu = fmt.Sprintf("%.4f", uploadThroughputTestResult.CpuAndRam.Cpu)
				ram = fmt.Sprintf("%d", uploadThroughputTestResult.CpuAndRam.Ram/1e6)
			}
			w.Write([]string{transferMode, bytes_MB, duration_ms, rate_MBps, rate_Mbps, cpu, ram})
			w.Flush()
		}
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

func testPing(logfilePrefix string, logfilePostfix string, serverHost string, serverPort uint, countSamples uint) {
	filename := testResultsDirectory + logfilePrefix + "-pingTest" + logfilePostfix + ".csv"
	contents := func(w *csv.Writer) {
		w.Write([]string{"ping average (ms)"})
		address := fmt.Sprintf("%s:%d", serverHost, serverPort)
		conn, err := net.Dial("udp", address)
		defer conn.Close()
		if err != nil {
			fmt.Printf("Some error %v", err)
			return
		}
		averagePingMicroSeconds := 0.0
		for i := uint(0); i < countSamples; i++ {
			dt := tests.Ping(conn)
			averagePingMicroSeconds += float64(dt.Microseconds())
			// fmt.Printf("Ping: %.3fms\n", float64(dt.Microseconds())/1000.0)
		}
		averagePingMicroSeconds /= float64(countSamples)
		fmt.Printf("Average Ping: %.3fms\n", averagePingMicroSeconds/1000.0)
		averagePingMicroSecondsString := fmt.Sprintf("%.3fms", averagePingMicroSeconds/1000.0)
		w.Write([]string{averagePingMicroSecondsString})
		w.Flush()
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}

func testJitter(logfilePrefix string, logfilePostfix string, serverHost string, serverPort uint, countDifferences uint) {
	filename := testResultsDirectory + logfilePrefix + "-jitterTest" + logfilePostfix + ".csv"
	contents := func(w *csv.Writer) {
		w.Write([]string{"Average Jitter (ms)"})
		address := fmt.Sprintf("%s:%d", serverHost, serverPort)
		conn, err := net.Dial("udp", address)
		defer conn.Close()
		if err != nil {
			fmt.Printf("Some error %v", err)
			return
		}
		averagePingMicroSeconds := 0.0
		dt1 := tests.Ping(conn)
		dt2 := time.Second
		for i := uint(0); i < countDifferences; i++ {
			dt2 = tests.Ping(conn)
			dtDiffMicroseconds := dt1.Microseconds() - dt2.Microseconds()
			if dtDiffMicroseconds < 0 {
				dtDiffMicroseconds *= -1
			}
			averagePingMicroSeconds += float64(dtDiffMicroseconds)
			dt1 = tests.Ping(conn)
		}
		averagePingMicroSeconds /= float64(countDifferences)
		fmt.Printf("Average Jitter: %.3fms\n", averagePingMicroSeconds/1000.0)
		averagePingMicroSecondsString := fmt.Sprintf("%.3fms", averagePingMicroSeconds/1000.0)
		w.Write([]string{averagePingMicroSecondsString})
		w.Flush()
	}
	createLogFile(filename, contents)
	fmt.Printf("\n")
}
