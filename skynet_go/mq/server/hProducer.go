package server

import (
	"net"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
)

func NewSHProducer(conn *Conn) *SHProducer {
	p := &SHProducer{}
	p.conn = conn
	p.Producer = mq.NewProducer(p)
	rpc.NewServer(p)
	return p
}

type SHProducer struct {
	rpc.ServerBase
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
