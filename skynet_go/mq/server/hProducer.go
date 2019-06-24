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

func (p *SHProducer) GetConn() mq.ConnI {
	if (p.client.GetState() & mq.ClientStateStart) > 0 {
		return p.client.GetConn()
	}

	return nil
}

func (p *SHProducer) GetDescribe() string {
	return p.client.GetDescribe()
}

func (p *SHProducer) Error(connI mq.ConnI) {
	p.client.Error(connI)
}
