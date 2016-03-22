package server

import (
	"net"
	"skp/skpPublic/skpSocket"
)

type SkpTcpServiceRoute struct {
	*skpSocket.SkpTcpService
	ManageRoute *SkpManageRoute
}

func SkpNewTcpServiceRoute() (this *SkpTcpServiceRoute) {
	tcpServiceRoute := new(SkpTcpServiceRoute)
	tcpServiceRoute.SkpTcpService = skpSocket.SkpNewTcpService(tcpServiceRoute)
	tcpServiceRoute.ManageRoute = SkpNewManageRoute(tcpServiceRoute)
	tcpServiceRoute.Manage = tcpServiceRoute.ManageRoute.SkpManage

	return tcpServiceRoute
}

func (this *SkpTcpServiceRoute) SkpAcceptConn(conn *net.TCPConn) {
	tcpSocketRoute := SkpNewTcpSocketRoute(conn, this)
	go tcpSocketRoute.SkpRun()
}
