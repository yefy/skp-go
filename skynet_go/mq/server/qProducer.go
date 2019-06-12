package server

import (
	"net"
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
	conn *Conn
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

func (q *SQProducer) GetTcp() (*net.TCPConn, int32, bool) {
	if q.conn == nil {
		if !q.GetConn() {
			return nil, 0, false
		}
	}

	if (q.conn.GetState() & mq.ConnStateStopSubscribe) > 0 {
		q.conn = nil
		return nil, 0, false
	}

	if (q.conn.GetState() & mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := q.conn.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (q *SQProducer) Error(tcpVersion int32) {
	q.conn.Error(tcpVersion)
	q.conn = nil
}
