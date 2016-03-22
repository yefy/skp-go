// goroutine project main.go
package main

import (
	"fmt"
	"github.com/yinheli/qqwry"
	"skp/skpMonitor/server"
)

func main() {
	q := qqwry.NewQQwry("qqwry.dat")
	q.Find("180.89.94.90")
	fmt.Printf("ip:%v, Country:%v, City:%v \n", q.Ip, q.Country, q.City)

	tcpService := server.SkpNewTcpServiceMonitor()
	tcpService.SkpLoadConf()
	tcpService.SkpService()
}
