package server

import (
	"net"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
	"strings"
	"sync"
)

func NewConn(server *Server, tcpConn *net.TCPConn) *Conn {
	c := &Conn{}
	c.server = server
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)
	c.Conn = mq.NewConn(tcpConn)
	rpc.NewServer(c)

	return c
}

type Conn struct {
	rpc.ServerB
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
	*mq.Conn
}

func (c *Conn) Start() {
	c.shConsumer.Start()
	c.shProducer.Start()
}

func (c *Conn) Stop() {
	c.shConsumer.Stop()
	c.shProducer.Stop()
}

func (c *Conn) Close() {
	c.Stop()
	c.RPC_GetServer().Stop(false)
	c.Conn.Close()
}

func (c *Conn) SendOnCloseAll() {
	c.RPC_GetServer().Send("OnCloseAll")
}

func (c *Conn) OnCloseAll() {
	c.Close()
}

func (c *Conn) Subscribe(topic string, tag string) {
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

func (c *Conn) IsSubscribe(tag string) bool {
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

func (c *Conn) IsSubscribeAll() bool {
	if c.tag == "*" {
		return true
	}

	return false
}
