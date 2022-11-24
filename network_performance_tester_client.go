package main

import (
	"github.com/jrcamenzuli/network-performance-tester-client/client"
	"github.com/jrcamenzuli/network-performance-tester-client/types"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

func main() {
	args := util.Args()
	var cfg types.Configuration
	util.ReadFile(&cfg, args.ConfigFile)
	util.ReadEnv(&cfg)
	client.RunClient(&cfg)
}
