package util

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
)

// ProcessInfo holds information about a discovered process
type ProcessInfo struct {
	PID  uint
	Name string
}

// FindProcessesByName finds all processes matching the given name
func FindProcessesByName(processName string) []ProcessInfo {
	var processes []ProcessInfo

	// Use pgrep to find processes by name
	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// No processes found or error - return empty slice
		return processes
	}

	pidStrings := strings.Fields(strings.TrimSpace(string(output)))
	for _, pidStr := range pidStrings {
		if pid, err := strconv.ParseUint(pidStr, 10, 32); err == nil {
			processes = append(processes, ProcessInfo{
				PID:  uint(pid),
				Name: processName,
			})
		}
	}

	return processes
}

// GetCPUandRAMForProcesses gets CPU and RAM usage for multiple processes by name
func GetCPUandRAMForProcesses(processNames []string) map[string]*model.CpuAndRam {
	result := make(map[string]*model.CpuAndRam)

	for _, processName := range processNames {
		processes := FindProcessesByName(processName)
		if len(processes) == 0 {
			// No processes found for this name
			result[processName] = &model.CpuAndRam{ProcessName: processName}
			continue
		}

		// Aggregate CPU and RAM usage for all processes with this name
		totalCPU := 0.0
		totalRAM := uint(0)
		processCount := 0

		for _, proc := range processes {
			cpuRam := GetCPUandRAM(proc.PID)
			if cpuRam != nil && (cpuRam.Cpu > 0 || cpuRam.Ram > 0) {
				totalCPU += cpuRam.Cpu
				totalRAM += cpuRam.Ram
				processCount++
			}
		}

		result[processName] = &model.CpuAndRam{
			Cpu:          totalCPU,
			Ram:          totalRAM,
			ProcessName:  processName,
			ProcessCount: processCount,
		}
	}

	return result
}

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
