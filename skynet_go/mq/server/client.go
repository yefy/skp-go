package server

import (
	"skp-go/skynet_go/encodes"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"strings"
	"sync"
	"time"
)

func NewClient(server *Server, connI mq.ConnI) *Client {
	c := &Client{}
	c.Client = mq.NewClient(c, connI)
	c.server = server
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)

	return c
}

type Client struct {
	*mq.Client
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
	return c.instance + "_s_Client"
}

func (c *Client) Error(connI mq.ConnI) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.Client.IsError(connI) {
		c.CloseSelf()
	}
}

func (c *Client) Start() {
	c.Stop()
	c.SetState(mq.ClientStateStart)
	c.shConsumer.Start()
	c.shProducer.Start()
}

func (c *Client) Stop() {
	c.shConsumer.Stop()
	c.shProducer.Stop()
}

func (c *Client) Close() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if (c.GetState() & (mq.ClientStateStopping | mq.ClientStateStop)) > 0 {
		return
	}

	c.AddState(mq.ClientStateStopping)
	c.CloseConn()
	c.Stop()
	c.RPC_GetServer().Stop(true)
	c.AddState(mq.ClientStateStop)
}

func (c *Client) CloseSelf() {
	c.server.OnMqDel(c)
	time.Sleep(time.Second)
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

	replyFunc := func(body interface{}) {
		sMqMsg, err := mq.ReplyMqMsg(c.harbor, mqMsg.GetPendingSeq(), mqMsg.GetEncode(), body)
		if err == nil {
			c.shProducer.RpcSend_OnWriteMqMsg(sMqMsg)
		}
	}

	if mqMsg.GetMethod() == mq.OnMqRegister {

		request := mq.RegisteRequest{}
		if err := encodes.DecodeBody(mqMsg.GetEncode(), mqMsg.GetBody(), &request); err != nil {
			c.CloseSelf()
			return
		}

		if request.Instance == "" {
			log.Err("request.Instance nil")
			c.CloseSelf()
			return
		}

		if request.Harbor > 0 {
			log.Err("request.Harbor = %d > 0", request.Harbor)
			c.CloseSelf()
			return
		}

		c.instance = request.Instance
		c.harbor = c.server.GetHarbor()
		c.Subscribe(request.Topic, request.Tag)

		c.SetState(mq.ClientStateStop)
		if !c.server.OnMqRegister(c) {
			log.Err("Instancer = %s : OnMqRegister falid", request.Instance)
			c.CloseSelf()
			return
		}

		reply := &mq.RegisterReply{}
		reply.Harbor = c.harbor
		replyFunc(reply)
		c.SetState(mq.ClientStateStart)
	} else if mqMsg.GetMethod() == mq.OnMqStopSubscribe {
		//先设置ClientStateStopSubscribe   qProducer.go 不会再向当前client发送msg
		c.AddState(mq.ClientStateStopSubscribe)
		reply := &mq.NilStruct{}
		replyFunc(reply)

	} else if mqMsg.GetMethod() == mq.OnMqClose {
		reply := &mq.NilStruct{}
		replyFunc(reply)
		c.CloseSelf()
	} else {
		log.Err("not class = %+v or method = %+v", mqMsg.GetClass(), mqMsg.GetMethod())
	}
}
