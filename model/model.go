package model

import (
	"fmt"
	"time"
)

type ThroughputType int64

const (
	TX            ThroughputType = 0
	RX                           = 1
	TX_FullDuplex                = 2
	RX_FullDuplex                = 3
)

func (e ThroughputType) String() string {
	switch e {
	case TX:
		return "TX Half Duplex"
	case RX:
		return "RX Half Duplex"
	case TX_FullDuplex:
		return "TX Full Duplex"
	case RX_FullDuplex:
		return "RX Full Duplex"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

type CpuAndRam struct {
	Pid uint
	Cpu float64
	Ram uint
}

type BurstTest struct {
	Duration    time.Duration
	FailureRate float64
	CpuAndRam   *CpuAndRam
}

type RateTest struct {
	FailureRate float64
	CpuAndRam   CpuAndRam
}

type Fn func(int) int

type ThroughputTest struct {
	Type                  ThroughputType
	CountBytesTransferred uint64
	DurationNanoseconds   uint64
	CpuAndRam             CpuAndRam
}

// Device Under Test Information
type DUT_Info struct {
	CPU_ModelName          string
	CPU_CoreCount          uint
	CPU_BaseClockFrequency uint // in Hz
	RAM_Total              uint // amount of bytes
}

func (e DUT_Info) String() string {
	return fmt.Sprintf("CPU_model_name: \"%s\", CPU_core_count: %d, CPU_base_clock_frequency: %.0fMHz, RAM_total: %.0fGB", e.CPU_ModelName, e.CPU_CoreCount, float64(e.CPU_BaseClockFrequency)/1e6, float64(e.RAM_Total)/1e9)
}
