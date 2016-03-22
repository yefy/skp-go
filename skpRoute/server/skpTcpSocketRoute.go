package server

import (
	"errors"
	"net"
	"skp/skpPublic/skpProtocol"
	"skp/skpPublic/skpSocket"
	"skp/skpPublic/skpUtility"
)

type SkpTcpSocketRoute struct {
	*skpSocket.SkpTcpSocket
	TcpServiceRoute *SkpTcpServiceRoute
}

func SkpNewTcpSocketRoute(conn *net.TCPConn, tcpServiceRoute *SkpTcpServiceRoute) (this *SkpTcpSocketRoute) {

	tcpSocketRoute := new(SkpTcpSocketRoute)
	tcpSocketRoute.SkpTcpSocket = skpSocket.SkpNewTcpSocket(tcpSocketRoute, conn, tcpServiceRoute.SkpTcpService)
	tcpSocketRoute.TcpServiceRoute = tcpServiceRoute

	return tcpSocketRoute
}

func (this *SkpTcpSocketRoute) SkpStart() {
	skpUtility.Println("SkpStartRoute")
}

func (this *SkpTcpSocketRoute) SkpData(data []byte) error {
	skpUtility.Println("SkpDataRoute")
	return this.SkpDataRoute(data)
}

func (this *SkpTcpSocketRoute) SkpTimeout() {
	//skpUtility.Println("SkpTimeoutRoute")
}

func (this *SkpTcpSocketRoute) SkpEnd() {
	skpUtility.Println("SkpEndRoute")
}

func (this *SkpTcpSocketRoute) SkpDataRoute(data []byte) error {
	skpUtility.Printf("**************************OrderRequst = %v \n", this.ProtocalHead.OrderRequst)

	if this.ProtocalHead.OrderType == skpProtocol.OrderTypeRequest {
		skpUtility.Printf("OrderType = %d invalid \n", this.ProtocalHead.OrderType)
		return errors.New("OrderType invalid")
	} else if this.ProtocalHead.OrderType == skpProtocol.OrderTypeResponse {
		switch this.ProtocalHead.OrderRequst {
		case skpProtocol.OrderMonitorLoadConfig:
			this.SkpAddMonitorIPList(data)
		default:
			skpUtility.Printf("Route data to Proxy = %d \n", this.ProtocalHead.ProxyID)

			this.Manage.SkpServerConn(skpProtocol.ConnTypeRoute, skpProtocol.ConnTypeProxy, uint64(this.ProtocalHead.ProxyID), data)
		}
	} else if this.ProtocalHead.OrderType == skpProtocol.OrderTypeServer {
		recv := this.ProtocalHead.Recv
		recvServerID := (recv >> 32) & 0xFFFFFFFF
		skpUtility.Printf("Route data to Local = %d \n", recvServerID)

		this.Manage.SkpServerConn(skpProtocol.ConnTypeRoute, skpProtocol.ConnTypeLocal, recvServerID, data)
	}

	return nil
}
