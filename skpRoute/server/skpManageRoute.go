package server

import (
	"net"
	"skp/skpPublic/skpSocket"
)

type SkpManageRoute struct {
	*skpSocket.SkpManage
	TcpServiceRoute *SkpTcpServiceRoute
}

func SkpNewManageRoute(tcpServiceRoute *SkpTcpServiceRoute) (this *SkpManageRoute) {
	manageRoute := new(SkpManageRoute)
	manageRoute.SkpManage = skpSocket.SkpNewManage(manageRoute)
	manageRoute.TcpServiceRoute = tcpServiceRoute
	return manageRoute
}

func (this *SkpManageRoute) SkpNewTcpSocket(conn *net.TCPConn, connType uint8) (*skpSocket.SkpTcpSocket, error) {
	tcpSocketRoute := SkpNewTcpSocketRoute(conn, this.TcpServiceRoute)
	return tcpSocketRoute.SkpTcpSocket, nil
}
