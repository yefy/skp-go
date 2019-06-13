package server

import (
	"net"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
	"strings"
	"sync"
)

func NewClient(server *Server, tcpConn *net.TCPConn) *Client {
	c := &Client{}
	c.server = server
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)
	c.Client = mq.NewClient(tcpConn)
	rpcU.NewServer(c)

	return c
}

type Client struct {
	rpcU.ServerB
	server   *Server
	harbor   int32
	instance string //topic_$$
	mutex    sync.Mutex

	topic  string
	tag    string
	tags   []string
	tagMap map[string]bool

	shConsumer *SHConsumer
	shProducer *SHProducer
	*mq.Client
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
	c.Stop()
	c.RPC_GetServer().Stop(false)
	c.Client.Close()
}

func (c *Client) SendOnCloseAll() {
	c.RPC_GetServer().Send("OnCloseAll")
}

func (c *Client) OnCloseAll() {
	c.Close()
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
