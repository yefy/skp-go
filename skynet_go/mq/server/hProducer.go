package server

import (
	"skp-go/skynet_go/mq"
)

func NewSHProducer(client *Client) *SHProducer {
	p := &SHProducer{}
	p.client = client
	p.Producer = mq.NewProducer(p)
	return p
}

type SHProducer struct {
	client *Client
	*mq.Producer
}

func (p *SHProducer) GetDescribe() string {
	return ""
}

func (p *SHProducer) GetConn() (mq.ConnI, int32, bool) {
	if (p.client.GetState() & mq.ClientStateStart) > 0 {
		connI, connVersion := p.client.GetConn()
		return connI, connVersion, true
	}

	return nil, 0, false
}

func (p *SHProducer) Error(connVersion int32) {
	p.client.Error(connVersion)
}
