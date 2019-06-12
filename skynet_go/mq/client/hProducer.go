package client

import (
	"net"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
)

func NewCHProducer(c *Client) *CHProducer {
	p := &CHProducer{}
	p.conn = c
	p.Producer = mq.NewProducer(p)
	rpc.NewServer(p)
	return p
}

type CHProducer struct {
	rpc.ServerB
	*mq.Producer
	conn *Client
}

func (p *CHProducer) GetTcp() (*net.TCPConn, int32, bool) {
	if (p.conn.GetState() & mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := p.conn.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (p *CHProducer) Error(tcpVersion int32) {
	p.conn.Error(tcpVersion)
}
