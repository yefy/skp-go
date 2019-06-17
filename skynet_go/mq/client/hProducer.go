package client

import (
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewCHProducer(cient *Client) *CHProducer {
	p := &CHProducer{}
	p.cient = cient
	p.Producer = mq.NewProducer(p)
	rpcU.NewServer(p)
	return p
}

type CHProducer struct {
	rpcU.ServerB
	*mq.Producer
	cient *Client
}

func (p *CHProducer) GetDescribe() string {
	return ""
}

func (p *CHProducer) GetConn() (conn.ConnI, int32, bool) {
	if (p.cient.GetState() & mq.ClientStateStart) > 0 {
		tcpConn, tcpVersion := p.cient.GetConn()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (p *CHProducer) Error(tcpVersion int32) {
	p.cient.Error(tcpVersion)
}
