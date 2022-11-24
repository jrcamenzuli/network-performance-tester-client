package util

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
)

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
