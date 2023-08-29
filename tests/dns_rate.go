package tests

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
	"github.com/miekg/dns"
)

func DnsRateTest(url string, testDuration time.Duration, desiredRequestsPerSecond int, pid uint, serverHost string, serverPort uint, transportProtocol string) model.RateTest {
	fmt.Printf("Sending %d DNS over %s requests per second for %s to %s:%d\n", desiredRequestsPerSecond, strings.ToUpper(transportProtocol), testDuration, serverHost, serverPort)
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

	c := dns.Client{Net: transportProtocol}
	c.Dial(fmt.Sprintf("%s:%d", serverHost, serverPort))
	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(url), dns.TypeA)

	time.Sleep(time.Duration(1.0 / float64(desiredRequestsPerSecond) * float64(time.Second)))
	for {

		duration = time.Since(tStart)
		if duration >= testDuration {
			break
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			resp, _, err := c.Exchange(&msg, fmt.Sprintf("%s:%d", serverHost, serverPort))
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if resp.Rcode != dns.RcodeSuccess {
				fmt.Printf("DNS query error: %s\n", dns.RcodeToString[resp.Rcode])
				return
			}

			// for _, ans := range resp.Answer {
			// 	if a, ok := ans.(*dns.A); ok {
			// 		fmt.Printf("IP address: %s\n", a.A.String())
			// 	}
			// }

			// fmt.Println("boop")
			if err == nil {
				countResponses++
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
