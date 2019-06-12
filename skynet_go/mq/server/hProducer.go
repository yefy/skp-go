package server

import (
	"net"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewSHProducer(conn *Conn) *SHProducer {
	p := &SHProducer{}
	p.conn = conn
	p.Producer = mq.NewProducer(p)
	rpcU.NewServer(p)
	return p
}

type SHProducer struct {
	rpcU.ServerB
	*mq.Producer
	conn *Conn
}

func (p *SHProducer) GetTcp() (*net.TCPConn, int32, bool) {
	if (p.conn.GetState() & mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := p.conn.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (p *SHProducer) Error(tcpVersion int32) {
	p.conn.Error(tcpVersion)
}
