package tests

import (
	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

// IdleStateOfProcess gets CPU/RAM usage for a single process by PID (deprecated)
func IdleStateOfProcess(pid uint) *model.CpuAndRam {
	return util.GetCPUandRAM(pid)
}

// IdleStateOfProcesses gets CPU/RAM usage for multiple processes by name
func IdleStateOfProcesses(processNames []string) map[string]*model.CpuAndRam {
	return util.GetCPUandRAMForProcesses(processNames)
}
