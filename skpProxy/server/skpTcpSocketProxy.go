package server

import (
	"net"
	"skp/skpPublic/skpProtocol"
	"skp/skpPublic/skpSocket"
	"skp/skpPublic/skpUtility"
)

type SkpTcpSocketProxy struct {
	*skpSocket.SkpTcpSocket
	TcpServiceProxy *SkpTcpServiceProxy
}

func SkpNewTcpSocketProxy(conn *net.TCPConn, tcpServiceProxy *SkpTcpServiceProxy) (this *SkpTcpSocketProxy) {

	tcpSocketProxy := new(SkpTcpSocketProxy)
	tcpSocketProxy.SkpTcpSocket = skpSocket.SkpNewTcpSocket(tcpSocketProxy, conn, tcpServiceProxy.SkpTcpService)
	tcpSocketProxy.TcpServiceProxy = tcpServiceProxy

	return tcpSocketProxy
}

func (this *SkpTcpSocketProxy) SkpStart() {
	skpUtility.Println("SkpStartProxy")
}

func (this *SkpTcpSocketProxy) SkpData(data []byte) error {
	//skpUtility.Println("SkpDataProxy")
	return this.SkpDataProxy(data)
}

func (this *SkpTcpSocketProxy) SkpTimeout() {
	//skpUtility.Println("SkpTimeoutProxy")
}

func (this *SkpTcpSocketProxy) SkpEnd() {
	skpUtility.Println("SkpEndProxy")
}

func (this *SkpTcpSocketProxy) SkpDataProxy(data []byte) error {

	if this.ProtocalHead.OrderType == skpProtocol.OrderTypeRequest {

		recv := this.ProtocalHead.Recv
		recvServerID := (recv >> 32) & 0xFFFFFFFF

		skpUtility.Printf("proxy data to serverID = %d \n", recvServerID)
		this.Manage.SkpServerConn(skpProtocol.ConnTypeProxy, skpProtocol.ConnTypeLocal, recvServerID, data)

	} else if this.ProtocalHead.OrderType == skpProtocol.OrderTypeResponse {
		switch this.ProtocalHead.OrderRequst {
		case skpProtocol.OrderMonitorLoadConfig:
			this.SkpAddMonitorIPList(data)
		default:
			skpUtility.Printf("proxy data to client = %d \n", this.ProtocalHead.Recv)
			this.Manage.SkpServer(this.ProtocalHead.Recv, data, this.ProtocalHead.AddrSocket)
		}
	}

	return nil
}
