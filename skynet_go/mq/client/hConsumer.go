package client

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
)

func NewCHConsumer(client *Client) *CHConsumer {
	c := &CHConsumer{}
	c.Consumer = mq.NewConsumer(c)
	c.client = client
	c.Start()
	return c
}

type CHConsumer struct {
	*mq.Consumer
	client *Client
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

	if rMqMsg.GetTopic() == "Mq" || rMqMsg.GetTag() == "*" {
		c.client.RegisterMqMsg(rMqMsg)
		return
	}

	if rMqMsg.GetTyp() == mq.TypeRespond {
		pendingMsg := c.client.GetPendingMsg(rMqMsg)
		if pendingMsg == nil {
			return
		}
		if pendingMsg.typ == mq.TypeCall {
			pendingMsg.pending <- rMqMsg
		}
	} else {
		rpcServer := c.client.GetRPCServer(rMqMsg)
		if rpcServer == nil {
			return
		}
		rpcServer.SendReq(rMqMsg.GetEncode(), rMqMsg.GetMethod(), rMqMsg.GetBody(), func(outStr string, err error) {
			log.Fatal("outStr = %+v, err = %+v", outStr, err)
			sMqMsg, err := mq.ReplyMqMsg(rMqMsg.GetHarbor(), rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), outStr)
			if err != nil {
				log.Err(err.Error())
				return
			}
			c.client.chProducer.RpcSend_OnWriteMqMsg(sMqMsg)
		})
	}
}
