package client

import (
	"net"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
)

func NewCHConsumer(conn *Client) *CHConsumer {
	c := &CHConsumer{}
	rpc.NewServer(c)
	c.conn = conn
	c.Consumer = mq.NewConsumer(c)
	c.Start()
	return c
}

type CHConsumer struct {
	rpc.ServerBase
	conn *Client
	*mq.Consumer
}

func (c *CHConsumer) GetTcp() (*net.TCPConn, int32, bool) {
	if (c.conn.GetState() & mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := c.conn.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (c *CHConsumer) Error(tcpVersion int32) {
	c.conn.Error(tcpVersion)
}

func (c *CHConsumer) DoMqMsg(rMqMsg *mq.MqMsg) {
	if rMqMsg.GetTyp() == mq.TypeRespond {
		pendingMsgI, ok := c.conn.pendingMap.Load(rMqMsg.GetPendingSeq())
		if !ok {
			log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
			return
		}
		c.conn.pendingMap.Delete(rMqMsg.GetPendingSeq())
		pendingMsg := pendingMsgI.(*PendingMsg)
		if pendingMsg.typ == mq.TypeCall {
			pendingMsg.pending <- rMqMsg
		}
	} else {
		c.conn.rpcEncode.SendReq(rMqMsg.GetEncode(), rMqMsg.GetMethod(), rMqMsg.GetBody(), func(outStr string, err error) {
			log.Fatal("outStr = %+v, err = %+v", outStr, err)
			sMqMsg, err := mq.ReplyMqMsg(rMqMsg.GetHarbor(), rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), outStr)
			if err != nil {
				log.Err(err.Error())
				return
			}
			c.conn.chProducer.SendWriteMqMsg(sMqMsg)
		})
	}
}
