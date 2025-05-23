package util

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
)

func GetCPUandRAM(pid uint) *model.CpuAndRam {
	pidString := fmt.Sprintf("%d", pid)
	cmd := exec.Command("bash", "-c", "top -l 2 | grep "+pidString+" | awk '{ printf(\"%s %d\\n\", $3, $8); }' | awk '{if(NR>1)print}'")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	raw := strings.Split(string(stdoutStderr), " ")
	if len(raw) == 2 {
		err = nil
		var cpu float64
		var ram uint64
		raw[1] = raw[1][:len(raw[1])-1]
		cpu, err = strconv.ParseFloat(raw[0], 32)
		ram, err = strconv.ParseUint(raw[1], 10, 64)
		return &model.CpuAndRam{Cpu: float64(cpu / 100.0), Ram: uint(ram)}
	}
	return &model.CpuAndRam{}
}

func GetSystemCPUUsage() float64 {
	// sysctl -n hw.ncpu
	// ps -A -o %cpu | awk '{s+=$1} END {print s "%"}'

	cmd1 := exec.Command("bash", "-c", "sysctl -n hw.ncpu")
	cmd2 := exec.Command("bash", "-c", "ps -A -o %cpu | awk '{s+=$1} END {print s}'")

	bytes1, err := cmd1.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	bytes2, err := cmd2.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	countCoresString := fmt.Sprintf("%s\n", bytes1)
	countCoresString = strings.TrimSpace(countCoresString)
	cpuUsageString := fmt.Sprintf("%s\n", bytes2)
	cpuUsageString = strings.TrimSpace(cpuUsageString)

	countCores, err := strconv.ParseUint(countCoresString, 10, 64)
	if err != nil {
		panic(err)
	}

	cpuUsage, err := strconv.ParseFloat(cpuUsageString, 64)
	if err != nil {
		panic(err)
	}

	cpuUsage /= 100.0

	return cpuUsage / float64(countCores)
}
