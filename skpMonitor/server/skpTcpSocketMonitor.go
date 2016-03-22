package server

import (
	"net"
	"skp/skpPublic/skpProtocol"
	"skp/skpPublic/skpSocket"
	"skp/skpPublic/skpUtility"
)

type SkpTcpSocketMonitor struct {
	*skpSocket.SkpTcpSocket
	TcpServiceMonitor *SkpTcpServiceMonitor
}

func SkpNewTcpSocketMonitor(conn *net.TCPConn, tcpServiceMonitor *SkpTcpServiceMonitor) (this *SkpTcpSocketMonitor) {

	tcpSocketMonitor := new(SkpTcpSocketMonitor)
	tcpSocketMonitor.SkpTcpSocket = skpSocket.SkpNewTcpSocket(tcpSocketMonitor, conn, tcpServiceMonitor.SkpTcpService)
	tcpSocketMonitor.TcpServiceMonitor = tcpServiceMonitor

	return tcpSocketMonitor
}

func (this *SkpTcpSocketMonitor) SkpStart() {
	skpUtility.Println("SkpStartMonitor")
}

func (this *SkpTcpSocketMonitor) SkpData(data []byte) error {
	return this.SkpDataMonitor(data)
}

func (this *SkpTcpSocketMonitor) SkpTimeout() {
	skpUtility.Println("SkpTimeoutMonitor")
}

func (this *SkpTcpSocketMonitor) SkpEnd() {
	skpUtility.Println("SkpEndtMonitor")
}

func (this *SkpTcpSocketMonitor) SkpDataMonitor(data []byte) error {
	skpUtility.Printf("**************************OrderRequst = %v \n", this.ProtocalHead.OrderRequst)

	if this.ProtocalHead.OrderType == skpProtocol.OrderTypeRequest {
		switch this.ProtocalHead.OrderRequst {
		case skpProtocol.OrderMonitorLoadConfig:

			addrData := make([]byte, 0, 10)
			addr := make([]byte, 0, 10)

			this.Manage.SkpGetServerAllAddr(&addr)

			var head skpProtocol.SkpProtocalHead
			skpProtocol.SkpPakegeHead(&head, uint32(len(addr)), uint64(this.Manage.ServerID), uint64(this.Manage.ServerID), skpProtocol.OrderTypeResponse, 0, this.ProtocalHead.OrderRequst, 0)

			h, _ := skpProtocol.SkpEncode(&head)
			addrData = append(addrData, h...)
			addrData = append(addrData, addr...)
			this.SkpWriteChan(addrData)
		}
	}

	return nil
}
