package server

import (
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewSHProducer(client *Client) *SHProducer {
	p := &SHProducer{}
	p.client = client
	p.Producer = mq.NewProducer(p)
	rpcU.NewServer(p)
	return p
}

type SHProducer struct {
	rpcU.ServerB
	*mq.Producer
	client *Client
}

func (p *SHProducer) GetTcp() (conn.ConnI, int32, bool) {
	if (p.client.GetState() & mq.ClientStateStart) > 0 {
		tcpConn, tcpVersion := p.client.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (p *SHProducer) Error(tcpVersion int32) {
	p.client.Error(tcpVersion)
}
