package server

import (
	"skp-go/skynet_go/mq"
)

func NewSQProducer(server *Server, topic string, tag string) *SQProducer {
	q := &SQProducer{}
	q.server = server
	q.topic = topic
	q.tag = tag
	q.key = q.topic + "_" + q.tag
	q.Producer = mq.NewProducer(q)
	return q
}

type SQProducer struct {
	server *Server
	topic  string
	tag    string
	key    string
	*mq.Producer
	client *Client
}

func (q *SQProducer) GetClient() bool {
	topicClients := q.server.topicClientsMap[q.topic]
	if topicClients != nil {
		for _, v := range topicClients.harborClient {
			if v.IsSubscribe(q.tag) {
				q.client = v
				return true
			}
		}
	}

	q.client = nil
	return false
}

func (q *SQProducer) GetDescribe() string {
	return ""
}

func (q *SQProducer) GetConn() (mq.ConnI, int32, bool) {
	if q.client == nil {
		if !q.GetClient() {
			return nil, 0, false
		}
	}

	if (q.client.GetState() & mq.ClientStateStopSubscribe) > 0 {
		q.client = nil
		return nil, 0, false
	}

	if (q.client.GetState() & mq.ClientStateStart) > 0 {
		connI, connVersion := q.client.GetConn()
		return connI, connVersion, true
	}

	return nil, 0, false
}

func (q *SQProducer) Error(connVersion int32) {
	q.client.Error(connVersion)
	q.client = nil
}
