package tests

import (
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
	"github.com/miekg/dns"
)

func test() {
	ips, err := net.LookupIP("google.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		fmt.Printf("google.com. IN A %s\n", ip.String())
	}
}

// todo
func DnsBurstTest(url string, burstSize int, pid uint, serverHost string, serverPort uint, transportProtocol string) model.BurstTest {
	return DnsBurstTestWithProcesses(url, burstSize, pid, serverHost, serverPort, transportProtocol, nil)
}

func DnsBurstTestWithProcesses(url string, burstSize int, pid uint, serverHost string, serverPort uint, transportProtocol string, processNames []string) model.BurstTest {
	countRequests := int32(0)
	countResponses := int32(0)
	var wg sync.WaitGroup

	fmt.Printf("Sending a burst of %d DNS over %s queries to %s:%d\n", burstSize, strings.ToUpper(transportProtocol), serverHost, serverPort)

	// Start monitoring processes if provided
	var processMonitoringDone sync.WaitGroup
	var processUsage model.ProcessCpuAndRam

	if len(processNames) > 0 {
		processMonitoringDone.Add(1)
		go func() {
			defer processMonitoringDone.Done()
			// Monitor for a reasonable burst test duration (5 seconds should be enough for most DNS bursts)
			processUsage = util.MonitorProcessesContinuously(processNames, 5*time.Second, 100*time.Millisecond)
		}()
	}

	tStart := time.Now()
	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			// Create a new client and message for each goroutine to avoid race conditions
			c := dns.Client{Net: transportProtocol}
			msg := dns.Msg{}
			msg.SetQuestion(dns.Fqdn(url), dns.TypeA)

			resp, _, err := c.Exchange(&msg, fmt.Sprintf("%s:%d", serverHost, serverPort))
			atomic.AddInt32(&countRequests, 1)

			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if resp.Rcode != dns.RcodeSuccess {
				fmt.Printf("DNS query error: %s\n", dns.RcodeToString[resp.Rcode])
				return
			}

			// Don't print IP address for successful queries - only count them
			atomic.AddInt32(&countResponses, 1)
		}(&wg)
	}
	cpuAndRam := util.GetCPUandRAM(pid)
	wg.Wait()
	tStop := time.Now()
	duration := tStop.Sub(tStart)

	// Stop process monitoring
	if len(processNames) > 0 {
		processMonitoringDone.Wait()
	}

	// Print summary
	successfulQueries := atomic.LoadInt32(&countResponses)
	totalQueries := atomic.LoadInt32(&countRequests)
	fmt.Printf("DNS Burst Test Summary: %d/%d queries successful in %dms\n", successfulQueries, totalQueries, duration.Milliseconds())

	failureRate := math.Max(0, 1.0-float64(countResponses)/float64(countRequests))
	return model.BurstTest{
		Duration:         duration,
		FailureRate:      failureRate,
		CpuAndRam:        cpuAndRam,
		ProcessCpuAndRam: processUsage,
	}
}
