package tests

import (
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

func IdleStateOfDevice() float64 {
	return util.GetSystemCPUUsage()
}
