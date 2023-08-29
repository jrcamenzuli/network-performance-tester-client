package tests

import (
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
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
	countRequests := 0
	countResponses := 0
	var wg sync.WaitGroup

	c := dns.Client{Net: transportProtocol}
	c.Dial(fmt.Sprintf("%s:%d", serverHost, serverPort))
	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(url), dns.TypeA)

	fmt.Printf("Sending a burst of %d DNS over %s queries to %s:%d\n", burstSize, strings.ToUpper(transportProtocol), serverHost, serverPort)
	tStart := time.Now()
	for i := 0; i <= burstSize; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			countRequests++

			resp, _, err := c.Exchange(&msg, fmt.Sprintf("%s:%d", serverHost, serverPort))
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if resp.Rcode != dns.RcodeSuccess {
				fmt.Printf("DNS query error: %s\n", dns.RcodeToString[resp.Rcode])
				return
			}

			for _, ans := range resp.Answer {
				if a, ok := ans.(*dns.A); ok {
					fmt.Printf("IP address: %s\n", a.A.String())
				}
			}

			if err == nil {
				countResponses++
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
