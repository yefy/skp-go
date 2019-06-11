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
	q.SHProducer = NewSHProducer(nil)
	q.SHProducerInterface = q
	return q
}

type SQProducer struct {
	server *Server
	topic  string
	tag    string
	key    string
	*SHProducer
}

func (q *SQProducer) GetConn() bool {
	topicConns := q.server.topicConnsMap[q.topic]
	if topicConns != nil {
		for _, v := range topicConns.harborConn {
			if v.IsSubscribe(q.tag) {
				q.conn = v
				return true
			}
		}
	}

	q.conn = nil
	return false
}

func (q *SQProducer) GetTcpConn() bool {
	if q.conn == nil {
		if !q.GetConn() {
			return false
		}
	}

	if (q.conn.GetState() & mq.ConnStateStopSubscribe) > 0 {
		q.conn = nil
		q.tcpConn = nil
		return false
	}

	return q.SHProducer.GetTcpConn()
}

func (q *SQProducer) Error() {
	q.SHProducer.Error()
	q.conn = nil
}
