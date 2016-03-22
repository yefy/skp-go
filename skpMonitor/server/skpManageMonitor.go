package server

import (
	"net"
	"skp/skpPublic/skpSocket"
)

type SkpManageMonitor struct {
	*skpSocket.SkpManage
	TcpServiceMonitor *SkpTcpServiceMonitor
}

func SkpNewManageMonitor(tcpServiceMonitor *SkpTcpServiceMonitor) (this *SkpManageMonitor) {
	manageMonitor := new(SkpManageMonitor)
	manageMonitor.SkpManage = skpSocket.SkpNewManage(manageMonitor)
	manageMonitor.TcpServiceMonitor = tcpServiceMonitor
	return manageMonitor
}

func (this *SkpManageMonitor) SkpNewTcpSocket(conn *net.TCPConn, connType uint8) (*skpSocket.SkpTcpSocket, error) {
	tcpSocketMonitor := SkpNewTcpSocketMonitor(conn, this.TcpServiceMonitor)
	return tcpSocketMonitor.SkpTcpSocket, nil
}
