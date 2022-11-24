package tests

import (
	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

func IdleStateOfProcess(pid uint) *model.CpuAndRam {
	return util.GetCPUandRAM(pid)
}
