package client

import (
	"skp-go/skynet_go/mq"
)

func NewCHProducer(client *Client) *CHProducer {
	p := &CHProducer{}
	p.Producer = mq.NewProducer(p)
	p.client = client
	return p
}

type CHProducer struct {
	*mq.Producer
	client *Client
}

func (p *CHProducer) GetConn() mq.ConnI {
	if (p.client.GetState() & mq.ClientStateStart) > 0 {
		return p.client.GetConn()
	}

	return nil
}

func (p *CHProducer) GetDescribe() string {
	return p.client.GetDescribe()
}

func (p *CHProducer) Error(connI mq.ConnI) {
	p.client.Error(connI)
}
