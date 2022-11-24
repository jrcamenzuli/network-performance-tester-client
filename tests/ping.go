package tests

import (
	"net"
	"time"
)

func Ping(conn net.Conn) time.Duration {
	tStart := time.Now()
	conn.Write([]byte{0})
	n := 0
	for n < 1 {
		n, _ = conn.Read([]byte{0})
		// fmt.Print(n)
	}
	tStop := time.Now()
	dt := tStop.Sub(tStart)
	return dt
}
