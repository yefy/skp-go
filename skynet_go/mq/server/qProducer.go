package server

import (
	"skp-go/skynet_go/mq"
)

func NewSQProducer(server *Server, topic string, tag string) *SQProducer {
	q := &SQProducer{}
	q.Producer = mq.NewProducer(q)
	q.server = server
	q.topic = topic
	q.tag = tag
	q.key = q.topic + "_" + q.tag
	return q
}

type SQProducer struct {
	*mq.Producer
	server *Server
	topic  string
	tag    string
	key    string
	client *Client
}

func (q *SQProducer) GetClient() bool {
	q.client = q.server.GetClient(q.topic, q.tag)
	return q.client != nil
}

func (q *SQProducer) GetConn() mq.ConnI {
	if q.client == nil {
		if !q.GetClient() {
			return nil
		}
	}

	if (q.client.GetState() & mq.ClientStateStopSubscribe) > 0 {
		q.client = nil
		return nil
	}

	if (q.client.GetState() & mq.ClientStateStart) > 0 {
		return q.client.GetConn()
	}

	return nil
}

func (q *SQProducer) GetDescribe() string {
	return q.key
}

func (q *SQProducer) Error(connI mq.ConnI) {
	q.client.Error(connI)
	q.client = nil
}
