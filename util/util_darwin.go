package util

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

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
			if cpuRam != nil {
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

// parseMemoryString parses memory strings with units (K, M, G) and returns bytes
func parseMemoryString(memStr string) uint64 {
	if len(memStr) == 0 {
		return 0
	}

	multiplier := uint64(1)
	numStr := memStr

	// Check for unit suffix
	lastChar := memStr[len(memStr)-1]
	switch lastChar {
	case 'K', 'k':
		multiplier = 1024
		numStr = memStr[:len(memStr)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		numStr = memStr[:len(memStr)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		numStr = memStr[:len(memStr)-1]
	}

	if ramVal, err := strconv.ParseFloat(numStr, 64); err == nil {
		return uint64(ramVal * float64(multiplier))
	}
	return 0
}

func GetCPUandRAM(pid uint) *model.CpuAndRam {
	pidString := fmt.Sprintf("%d", pid)
	cmd := exec.Command("bash", "-c", "top -l 2 | grep "+pidString+" | awk '{ printf(\"%s %s\\n\", $3, $8); }' | awk '{if(NR>1)print}'")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	raw := strings.Split(strings.TrimSpace(string(stdoutStderr)), " ")
	if len(raw) != 2 {
		return &model.CpuAndRam{}
	}

	// Parse CPU percentage
	cpu, err := strconv.ParseFloat(raw[0], 64)
	if err != nil {
		cpu = 0.0
	}

	// Parse memory with unit suffixes
	ram := parseMemoryString(strings.TrimSpace(raw[1]))

	return &model.CpuAndRam{
		Cpu: cpu / 100.0, // Convert percentage to decimal
		Ram: uint(ram),
	}
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

// MonitorProcessesContinuously monitors processes continuously and returns average CPU and maximum RAM usage
func MonitorProcessesContinuously(processNames []string, duration time.Duration, sampleInterval time.Duration) model.ProcessCpuAndRam {
	result := make(model.ProcessCpuAndRam)

	// Initialize result map
	for _, processName := range processNames {
		result[processName] = &model.CpuAndRam{
			ProcessName: processName,
			Cpu:         0.0,
			Ram:         0,
		}
	}

	if len(processNames) == 0 {
		return result
	}

	var wg sync.WaitGroup
	mutex := &sync.Mutex{}

	// Track samples and max RAM for each process
	sampleCounts := make(map[string]int)
	maxRAM := make(map[string]uint)
	totalCPU := make(map[string]float64)

	for _, processName := range processNames {
		sampleCounts[processName] = 0
		maxRAM[processName] = 0
		totalCPU[processName] = 0.0
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		startTime := time.Now()

		for time.Since(startTime) < duration {
			currentUsage := GetCPUandRAMForProcesses(processNames)

			mutex.Lock()
			for processName, usage := range currentUsage {
				if usage != nil {
					totalCPU[processName] += usage.Cpu
					sampleCounts[processName]++

					// Track maximum RAM usage
					if usage.Ram > maxRAM[processName] {
						maxRAM[processName] = usage.Ram
					}

					// Update process count
					result[processName].ProcessCount = usage.ProcessCount
				}
			}
			mutex.Unlock()

			time.Sleep(sampleInterval)
		}
	}()

	wg.Wait()

	// Calculate averages
	mutex.Lock()
	for processName := range result {
		if sampleCounts[processName] > 0 {
			result[processName].Cpu = totalCPU[processName] / float64(sampleCounts[processName])
		}
		result[processName].Ram = maxRAM[processName]
	}
	mutex.Unlock()

	return result
}
