// goroutine project main.go
package main

import (
	"skp/skpProxy/server"
	"skp/skpPublic/skpProtocol"
)

func main() {
	tcpService := server.SkpNewTcpServiceProxy()
	tcpService.SkpLoadConf(skpProtocol.ConnTypeProxy)
	tcpService.SkpService()
}
