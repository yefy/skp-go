package client

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewCHConsumer(cient *Client) *CHConsumer {
	c := &CHConsumer{}
	rpcU.NewServer(c)
	c.cient = cient
	c.Consumer = mq.NewConsumer(c)
	c.Start()
	return c
}

type CHConsumer struct {
	rpcU.ServerB
	cient *Client
	*mq.Consumer
}

func (c *CHConsumer) GetTcp() (conn.ConnI, int32, bool) {
	if (c.cient.GetState() & mq.ClientStateStart) > 0 {
		tcpConn, tcpVersion := c.cient.GetTcp()
		return tcpConn, tcpVersion, true
	}

	return nil, 0, false
}

func (c *CHConsumer) Error(tcpVersion int32) {
	c.cient.Error(tcpVersion)
}

func (c *CHConsumer) DoMqMsg(rMqMsg *mq.MqMsg) {
	if rMqMsg.GetTyp() == mq.TypeRespond {
		pendingMsgI, ok := c.cient.pendingMap.Load(rMqMsg.GetPendingSeq())
		if !ok {
			log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
			return
		}
		c.cient.pendingMap.Delete(rMqMsg.GetPendingSeq())
		pendingMsg := pendingMsgI.(*PendingMsg)
		if pendingMsg.typ == mq.TypeCall {
			pendingMsg.pending <- rMqMsg
		}
	} else {
		rpcServer := c.cient.rpcEMap[rMqMsg.GetClass()]
		if rpcServer == nil {
			log.Fatal("not rMqMsg.GetClass() = %+v", rMqMsg.GetClass())
			return
		}
		rpcServer.SendReq(rMqMsg.GetEncode(), rMqMsg.GetMethod(), rMqMsg.GetBody(), func(outStr string, err error) {
			log.Fatal("outStr = %+v, err = %+v", outStr, err)
			sMqMsg, err := mq.ReplyMqMsg(rMqMsg.GetHarbor(), rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), outStr)
			if err != nil {
				log.Err(err.Error())
				return
			}
			c.cient.chProducer.SendWriteMqMsg(sMqMsg)
		})
	}
}
