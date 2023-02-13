package tests

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

func HttpRateTest(url string, testDuration time.Duration, desiredRequestsPerSecond int, pid uint, isHttps bool) model.RateTest {
	protocol := ""
	if isHttps {
		protocol = "HTTPS"
	} else {
		protocol = "HTTP"
	}

	fmt.Printf("Sending %d %s requests per second for %s to %s\n", desiredRequestsPerSecond, protocol, testDuration, url)

	countRequests := 0
	countResponses := 0
	countSamples := 0
	var duration time.Duration
	cpuAndRam := model.CpuAndRam{Pid: pid}
	Kp := 2.0
	Ki := 1.2
	Kd := 0.001
	integral := 0.0
	previous_error := 0.0

	sumActualRequestsPerSecond := 0.0

	var wg sync.WaitGroup

	tStart := time.Now()
	tLast := tStart

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		defer func(cpuAndRam *model.CpuAndRam) {
			cpuAndRam.Cpu /= float64(countSamples)
		}(&cpuAndRam)
		for time.Since(tStart) < testDuration {
			a := util.GetCPUandRAM(pid)
			countSamples++
			cpuAndRam.Cpu += a.Cpu
		}
		cpuAndRam.Ram = util.GetCPUandRAM(pid).Ram
	}(&wg)

	time.Sleep(time.Duration(1.0 / float64(desiredRequestsPerSecond) * float64(time.Second)))
	for {

		duration = time.Since(tStart)
		if duration >= testDuration {
			break
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			resp, err := http.Get(url)
			// fmt.Println("boop")
			if err == nil {
				countResponses++
				defer resp.Body.Close()
			} else {
				return
			}
		}(&wg)
		countRequests++

		dt := time.Since(tLast)
		tLast = time.Now()
		sumActualRequestsPerSecond += float64(countRequests) / float64(duration.Seconds())
		averageActualRequestsPerSecond := sumActualRequestsPerSecond / float64(countRequests)

		error_ := 1.0/float64(desiredRequestsPerSecond) - 1.0/averageActualRequestsPerSecond
		proportional := error_
		integral = integral + error_*dt.Seconds()
		derivative := (error_ - previous_error) / dt.Seconds()
		output := Kp*proportional + Ki*integral + Kd*derivative
		if output == math.NaN() || math.IsInf(output, 0) {
			output = 0.0
		}
		previous_error = error_

		fmt.Printf("desiredRequestsPerSecond:%d, averageActualRequestsPerSecond:%f, output:%f\n", desiredRequestsPerSecond, averageActualRequestsPerSecond, output)
		time.Sleep(time.Duration(1.0/float64(desiredRequestsPerSecond)*float64(time.Second)) + time.Duration(output*float64(time.Second)))
	}

	wg.Wait()

	failureRate := math.Max(0, 1.0-float64(countResponses)/float64(countRequests))
	return model.RateTest{FailureRate: failureRate, CpuAndRam: cpuAndRam}
}
