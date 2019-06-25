package server

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
)

func NewSHConsumer(client *Client) *SHConsumer {
	c := &SHConsumer{}
	c.Consumer = mq.NewConsumer(c)
	c.client = client
	return c
}

type SHConsumer struct {
	*mq.Consumer
	client *Client
}

func (c *SHConsumer) GetConn() mq.ConnI {
	if (c.client.GetState() & mq.ClientStateStart) > 0 {
		return c.client.GetConn()
	}

	return nil
}

func (c *SHConsumer) GetDescribe() string {
	return c.client.GetDescribe()
}

func (c *SHConsumer) Error(connI mq.ConnI) {
	c.client.Error(connI)
}

func (c *SHConsumer) DoMqMsg(mqMsg *mq.MqMsg) {
	if c.client.harbor == 0 {
		log.Err("%s, harbor = %d", c.GetDescribe(), c.client.harbor)
		if mqMsg == nil {
			log.Err("%s, mqMsg == nil", c.GetDescribe())
			c.client.CloseSelf()
			return
		} else if mqMsg.GetTyp() == mq.TypeRespond {
			log.Err("%s, mqMsg.GetTyp() == mq.TypeRespond", c.GetDescribe())
			c.client.CloseSelf()
			return
		} else if mqMsg.GetTopic() != "Mq" {
			log.Err("%s, mqMsg.GetTopic() != Mq", c.GetDescribe())
			c.client.CloseSelf()
			return
		}
	}

	if mqMsg.GetTopic() == "Mq" {
		if mqMsg.GetTag() != "*" {
			log.Err("%s, mqMsg.GetTag() != *", c.GetDescribe())
			c.client.CloseSelf()
			return
		}
		c.client.RegisterMqMsg(mqMsg)
		return
	}

	if mqMsg.GetTyp() == mq.TypeRespond {
		harborClient := c.client.server.GetHarborClient(mqMsg)
		if harborClient == nil {
			return
		}
		harborClient.shProducer.RpcSend_OnWriteMqMsg(mqMsg)
	} else {
		q := c.client.server.GetSQProducer(mqMsg)
		q.RpcSend_OnWriteMqMsg(mqMsg)
	}
}
