package server

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
)

func NewSHConsumer(client *Client) *SHConsumer {
	c := &SHConsumer{}
	c.client = client
	c.Consumer = mq.NewConsumer(c)
	rpcU.NewServer(c)
	return c
}

type SHConsumer struct {
	rpcU.ServerB
	client *Client
	*mq.Consumer
}

func (c *SHConsumer) GetDescribe() string {
	return ""
}

func (c *SHConsumer) GetConn() (mq.ConnI, int32, bool) {
	if (c.client.GetState() & mq.ClientStateStart) > 0 {
		connI, connVersion := c.client.GetConn()
		return connI, connVersion, true
	}

	return nil, 0, false
}

func (c *SHConsumer) Error(connVersion int32) {
	c.client.Error(connVersion)
}

func (c *SHConsumer) DoMqMsg(mqMsg *mq.MqMsg) {
	if mqMsg == nil && c.client.harbor == 0 {
		c.client.server.ClientError(c.client)
		return
	}

	if mqMsg.GetTyp() == mq.TypeRespond {
		harbor := mqMsg.GetHarbor()
		harborClientI, ok := c.client.server.harborClient.Load(harbor)
		if ok {
			harborClient := harborClientI.(*Client)
			harborClient.shProducer.SendWriteMqMsg(mqMsg)
		} else {
			//这里需要保存mqMsg用于排查问题
			log.Fatal("not harbor = %d", harbor)
		}
	} else {
		if mqMsg.GetTopic() == "Mq" {
			c.client.server.OnRegisterMqMsg(c.client, mqMsg)
		} else {
			if c.client.harbor == 0 {
				c.client.server.ClientError(c.client)
				return
			}

			key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
			q := c.client.server.topicTag[key]
			if q == nil {
				log.Fatal("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
				q = NewSQProducer(c.client.server, mqMsg.GetTopic(), mqMsg.GetTag())
				c.client.server.topicTag[key] = q
			}
			q.SendWriteMqMsg(mqMsg)
		}
	}
}
