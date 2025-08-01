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
	cpuAndRam := util.GetCPUandRAM(pid)
	wg.Wait()
	tStop := time.Now()
	duration := tStop.Sub(tStart)
	failureRate := math.Max(0, 1.0-float64(countResponses)/float64(countRequests))
	return model.BurstTest{Duration: duration, FailureRate: failureRate, CpuAndRam: cpuAndRam}
}
