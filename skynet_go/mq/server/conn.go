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
	c.SetTcpConn(tcpConn)
	rpc.NewServer(c)

	return c
}

type Conn struct {
	rpc.ServerBase
	server     *Server
	harbor     int32
	instance   string //topic_$$
	mutex      sync.Mutex
	tcpConn    *net.TCPConn
	tcpVersion int32
	state      int32

	topic  string
	tag    string
	tags   []string
	tagMap map[string]bool

	shConsumer *SHConsumer
	shProducer *SHProducer
}

func (c *Conn) SendError(tcpVersion int32) {
	c.RPC_GetServer().Send("OnError", tcpVersion)
}

func (c *Conn) OnError(tcpVersion int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if tcpVersion != c.tcpVersion {
		return
	}
	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
	c.tcpVersion++
	c.state = mq.ConnStateErr
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

	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
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

func (c *Conn) GetState() int32 {
	return c.state
}

func (c *Conn) GetTcpConn() (*net.TCPConn, int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn, c.tcpVersion
}

func (c *Conn) SetTcpConn(tcpConn *net.TCPConn) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}

	c.tcpConn = tcpConn
	c.tcpVersion++
	c.state = mq.ConnStateStart
}
