package server

import (
	"skp-go/skynet_go/mq"
)

func NewSHProducer(client *Client) *SHProducer {
	p := &SHProducer{}
	p.Producer = mq.NewProducer(p)
	p.client = client
	return p
}

type SHProducer struct {
	*mq.Producer
	client *Client
}

func (p *SHProducer) GetConn() mq.ConnI {
	if (p.client.GetState() & mq.ClientStateStart) > 0 {
		return p.client.GetConn()
	}

	return nil
}

func (p *SHProducer) GetDescribe() string {
	return p.client.GetDescribe() + "_SHProducer"
}

func (p *SHProducer) Error(connI mq.ConnI) {
	p.client.Error(connI)
}
