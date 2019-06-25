package server

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
	"strings"
	"sync"
)

func NewClient(server *Server, connI mq.ConnI) *Client {
	c := &Client{}
	c.Client = mq.NewClient(connI)
	rpcU.NewServer(c)
	c.server = server
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)

	return c
}

type Client struct {
	*mq.Client
	rpcU.ServerB
	server     *Server
	shConsumer *SHConsumer
	shProducer *SHProducer
	harbor     int32
	instance   string //topic_$$
	mutex      sync.Mutex
	topic      string
	tag        string
	tags       []string
	tagMap     map[string]bool
}

func (c *Client) GetDescribe() string {
	return c.instance
}

func (c *Client) Start() {
	c.shConsumer.Start()
	c.shProducer.Start()
}

func (c *Client) Stop() {
	c.shConsumer.Stop()
	c.shProducer.Stop()
}

func (c *Client) Close() {
	c.SetState(mq.ClientStateStopping)
	c.Stop()
	c.RPC_GetServer().Stop(false)
	c.CloseConn()
	c.SetState(mq.ClientStateStop)
}

func (c *Client) CloseSelf() {
	go c.Close()
}

func (c *Client) Subscribe(topic string, tag string) {
	c.topic = strings.Trim(topic, "\t\n ")
	c.tag = strings.Trim(tag, "\t\n ")
	if c.tag == "*" {
		return
	}

	c.tags = strings.Split(c.tag, "||")
	for _, v := range c.tags {
		c.tagMap[v] = true
	}
}

func (c *Client) IsSubscribe(tag string) bool {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	log.Fatal("IsSubscribe tag = %s", tag)
	if c.tag == "*" {
		return true
	}

	if c.tagMap[tag] {
		return true
	}

	return false
}

func (c *Client) IsSubscribeAll() bool {
	if c.tag == "*" {
		return true
	}

	return false
}

func (c *Client) RegisterMqMsg(mqMsg *mq.MqMsg) {
	if mqMsg.GetClass() != "Mq" {
		log.Err("mqMsg.GetClass() != Mq")
		return
	}

	if mqMsg.GetMethod() == mq.OnMqRegister {
		c.server.OnClientRegister(c, mqMsg)

	} else if mqMsg.GetMethod() == mq.OnMqStopSubscribe {
		c.SetState(mq.ClientStateStart | mq.ClientStateStopSubscribe)

	} else if mqMsg.GetMethod() == mq.OnMqClose {
		//清除client
		c.CloseSelf()

	} else {
		log.Err("not class = %+v or method = %+v", mqMsg.GetClass(), mqMsg.GetMethod())
	}
}
