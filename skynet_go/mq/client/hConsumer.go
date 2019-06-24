package client

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewCHConsumer(client *Client) *CHConsumer {
	c := &CHConsumer{}
	rpcU.NewServer(c)
	c.client = client
	c.Consumer = mq.NewConsumer(c)
	c.Start()
	return c
}

type CHConsumer struct {
	rpcU.ServerB
	client *Client
	*mq.Consumer
}

func (c *CHConsumer) GetConn() mq.ConnI {
	if (c.client.GetState() & mq.ClientStateStart) > 0 {
		return c.client.GetConn()
	}

	return nil
}

func (c *CHConsumer) GetDescribe() string {
	return c.client.GetDescribe()
}

func (c *CHConsumer) Error(connI mq.ConnI) {
	c.client.Error(connI)
}

func (c *CHConsumer) DoMqMsg(rMqMsg *mq.MqMsg) {
	if rMqMsg == nil {
		return
	}

	if rMqMsg.GetTyp() == mq.TypeRespond {
		pendingMsgI, ok := c.client.pendingMap.Load(rMqMsg.GetPendingSeq())
		if !ok {
			log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
			return
		}
		c.client.pendingMap.Delete(rMqMsg.GetPendingSeq())
		pendingMsg := pendingMsgI.(*PendingMsg)
		if pendingMsg.typ == mq.TypeCall {
			pendingMsg.pending <- rMqMsg
		}
	} else {
		//___yefy 有序待添加
		rpcServer := c.client.rpcEMap[rMqMsg.GetClass()]
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
			c.client.chProducer.SendWriteMqMsg(sMqMsg)
		})
	}
}
