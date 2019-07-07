package client

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
)

func NewCHProducer(client *Client) *CHProducer {
	p := &CHProducer{}
	log.Fatal("NewCHProducer 000000000")
	p.Producer = mq.NewProducer(p)
	log.Fatal("NewCHProducer 1111111111")
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
	return p.client.GetDescribe() + "_CHProducer"
}

func (p *CHProducer) Error(connI mq.ConnI) {
	p.client.Error(connI)
}
