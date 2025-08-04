package util

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
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

// FindProcessesByName finds all processes matching the given name on Windows
func FindProcessesByName(processName string) []ProcessInfo {
	var processes []ProcessInfo

	// Use PowerShell to find processes by name
	cmd := exec.Command("powershell", "-nologo", "-noprofile", "-command",
		fmt.Sprintf("Get-Process | Where-Object {$_.ProcessName -like '*%s*'} | Select-Object Id", processName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// No processes found or error - return empty slice
		return processes
	}

	// Parse the PowerShell output to extract PIDs
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Id") || strings.Contains(line, "--") {
			continue
		}

		if pid, err := strconv.ParseUint(line, 10, 32); err == nil {
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
	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "$result=get-wmiobject -class Win32_PerfFormattedData_PerfProc_Process | where {$_.IDProcess -eq \"%d\"} | select percentprocessortime, workingsetprivate | select @{Name=\"CPU\";Expression={($_.percentprocessortime / (Get-WMIObject Win32_ComputerSystem).NumberOfLogicalProcessors)}},@{Name=\"MEM\";Expression=\"workingsetprivate\"} -first 1\n", pid)
		fmt.Fprintf(stdin, "Write-Output $result\n")
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`((\d+.\d+)|(\d+))\s+\d+`)
	rows := re.FindAllString(string(out), -1)
	if len(rows) <= 0 {
		return &model.CpuAndRam{}
	}
	row := rows[0]
	row = strings.TrimSpace(row)
	values := strings.Split(row, " ")
	cpuUsage := float64(0)
	cpuUsage, _ = strconv.ParseFloat(values[0], 64)
	memoryUsage := uint64(0)
	memoryUsage, _ = strconv.ParseUint(values[1], 10, 64)
	cpuAndRam := &model.CpuAndRam{Pid: pid, Cpu: (float64(cpuUsage) / 100.0), Ram: uint(memoryUsage)}
	return cpuAndRam
}

func GetSystemCPUUsage() float64 {
	// (Get-CimInstance Win32_ComputerSystem).NumberOfLogicalProcessors
	// Get-WmiObject Win32_Processor | Select LoadPercentage | Format-List

	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "Get-WmiObject Win32_Processor | Select LoadPercentage | Format-List\n")
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`\d+`)
	match := re.FindAllString(string(out), -1)
	val, _ := strconv.ParseFloat(match[len(match)-1], 64)
	return val / 100.0
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
