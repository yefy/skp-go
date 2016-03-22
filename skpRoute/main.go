// goroutine project main.go
package main

import (
	"skp/skpPublic/skpProtocol"
	"skp/skpRoute/server"
)

func main() {
	tcpService := server.SkpNewTcpServiceRoute()
	tcpService.SkpLoadConf(skpProtocol.ConnTypeRoute)
	tcpService.SkpService()
}
