package tests

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

func HttpBurstTest(url string, burstSize int, pid uint, isHttps bool) model.BurstTest {
	return HttpBurstTestWithProcesses(url, burstSize, pid, isHttps, nil)
}

func HttpBurstTestWithProcesses(url string, burstSize int, pid uint, isHttps bool, processNames []string) model.BurstTest {
	protocol := ""
	if isHttps {
		protocol = "HTTPS"
	} else {
		protocol = "HTTP"
	}
	countRequests := 0
	countResponses := 0
	var wg sync.WaitGroup

	fmt.Printf("Sending a burst of %d %s requests to %s\n", burstSize, protocol, url)
	client := util.CreateHTTPSClient()

	tStart := time.Now()

	// Start monitoring processes if provided
	var processMonitoringDone sync.WaitGroup
	var processUsage model.ProcessCpuAndRam

	if len(processNames) > 0 {
		processMonitoringDone.Add(1)
		go func() {
			defer processMonitoringDone.Done()
			// Monitor for a reasonable burst test duration (5 seconds should be enough for most bursts)
			processUsage = util.MonitorProcessesContinuously(processNames, 5*time.Second, 100*time.Millisecond)
		}()
	}
	for i := 0; i <= burstSize; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			countRequests++
			resp, err := client.Get(url)
			if err == nil {
				countResponses++
				defer resp.Body.Close()
			} else {
				return
			}
		}(&wg)
	}

	// Legacy single-process monitoring (backward compatibility)
	cpuAndRam := util.GetCPUandRAM(pid)
	wg.Wait()
	tStop := time.Now()
	duration := tStop.Sub(tStart)

	// Stop process monitoring
	if len(processNames) > 0 {
		processMonitoringDone.Wait()
	}

	failureRate := math.Max(0, 1.0-float64(countResponses)/float64(countRequests))
	return model.BurstTest{
		Duration:         duration,
		FailureRate:      failureRate,
		CpuAndRam:        cpuAndRam,
		ProcessCpuAndRam: processUsage,
	}
}
