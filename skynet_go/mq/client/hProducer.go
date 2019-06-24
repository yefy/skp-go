package client

import (
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewCHProducer(client *Client) *CHProducer {
	p := &CHProducer{}
	p.client = client
	p.Producer = mq.NewProducer(p)
	rpcU.NewServer(p)
	return p
}

type CHProducer struct {
	rpcU.ServerB
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
