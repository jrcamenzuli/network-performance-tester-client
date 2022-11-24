package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jrcamenzuli/network-performance-tester-client/types"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

func Args() types.ProgramArgs {
	configPtr := flag.String("config", "config.yml", "")
	// if len(*postfixPtr) > 0 {
	// 	(*postfixPtr) = "-" + *postfixPtr
	// }
	flag.Parse()
	fmt.Println()
	fmt.Println("config:", *configPtr)
	fmt.Println()

	if _, err := os.Stat("test-results"); os.IsNotExist(err) {
		err := os.Mkdir("test-results", 0700)
		if err != nil {
			fmt.Println("Could not create test-results directory")
		}
	}

	return types.ProgramArgs{ConfigFile: *configPtr}
}

func Fib(n int) int {
	if n <= 1 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func ReadFile(cfg *types.Configuration, configFile string) {
	f, err := os.Open(configFile)
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func ReadEnv(cfg *types.Configuration) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}

func PrettifyStruct(o interface{}) string {
	out, err := json.MarshalIndent(o, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(out)
}
