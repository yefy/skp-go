package server

import (
	"net"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
)

func NewSHConsumer(conn *Conn) *SHConsumer {
	c := &SHConsumer{}
	c.conn = conn
	c.Consumer = mq.NewConsumer(c)
	rpc.NewServer(c)
	return c
}

type SHConsumer struct {
	rpc.ServerBase
	conn *Conn
	*mq.Consumer
}

func (c *SHConsumer) GetTcp() (*net.TCPConn, int32, bool) {
	if (c.conn.GetState() & mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := c.conn.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (c *SHConsumer) Error(tcpVersion int32) {
	c.conn.Error(tcpVersion)
}

func (c *SHConsumer) DoMqMsg(mqMsg *mq.MqMsg) {
	if mqMsg.GetTyp() == mq.TypeRespond {
		harbor := mqMsg.GetHarbor()
		harborConnI, ok := c.conn.server.harborConn.Load(harbor)
		if ok {
			harborConn := harborConnI.(*Conn)
			harborConn.shProducer.SendWriteMqMsg(mqMsg)
		} else {
			//这里需要保存mqMsg用于排查问题
			log.Fatal("not harbor = %d", harbor)
		}
	} else {
		key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
		q := c.conn.server.topicTag[key]
		if q == nil {
			log.Fatal("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
			q = NewSQProducer(c.conn.server, mqMsg.GetTopic(), mqMsg.GetTag())
			c.conn.server.topicTag[key] = q
		}
		q.SendWriteMqMsg(mqMsg)
	}
}
