package main

import (
	"fmt"
	"net"
)

func main() {
	ips, err := net.LookupIP("server")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, ip := range ips {
		fmt.Println("IP:", ip)
	}
}
