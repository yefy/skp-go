package server

import (
	"net"
	"skp/skpPublic/skpSocket"
)

type SkpTcpServiceProxy struct {
	*skpSocket.SkpTcpService
	ManageProxy *SkpManageProxy
}

func SkpNewTcpServiceProxy() (this *SkpTcpServiceProxy) {
	tcpServiceProxy := new(SkpTcpServiceProxy)
	tcpServiceProxy.SkpTcpService = skpSocket.SkpNewTcpService(tcpServiceProxy)
	tcpServiceProxy.ManageProxy = SkpNewManageProxy(tcpServiceProxy)
	tcpServiceProxy.Manage = tcpServiceProxy.ManageProxy.SkpManage

	return tcpServiceProxy
}

func (this *SkpTcpServiceProxy) SkpAcceptConn(conn *net.TCPConn) {
	tcpSocketProxy := SkpNewTcpSocketProxy(conn, this)
	go tcpSocketProxy.SkpRun()
}
