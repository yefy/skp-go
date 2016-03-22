package server

import (
	"net"
	"skp/skpPublic/skpSocket"
)

type SkpManageProxy struct {
	*skpSocket.SkpManage
	TcpServiceProxy *SkpTcpServiceProxy
}

func SkpNewManageProxy(tcpServiceProxy *SkpTcpServiceProxy) (this *SkpManageProxy) {
	manageProxy := new(SkpManageProxy)
	manageProxy.SkpManage = skpSocket.SkpNewManage(manageProxy)
	manageProxy.TcpServiceProxy = tcpServiceProxy
	return manageProxy
}

func (this *SkpManageProxy) SkpNewTcpSocket(conn *net.TCPConn, connType uint8) (*skpSocket.SkpTcpSocket, error) {
	tcpSocketProxy := SkpNewTcpSocketProxy(conn, this.TcpServiceProxy)
	return tcpSocketProxy.SkpTcpSocket, nil
}
