package mq

import (
	"net"
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
)

func NewConn(tcpConn *net.TCPConn) *Conn {
	c := &Conn{}
	rpcU.NewServer(c)
	if tcpConn != nil {
		c.SetTcp(tcpConn)
	}

	return c
}

type Conn struct {
	rpcU.ServerB
	mutex      sync.Mutex
	tcpConn    *net.TCPConn
	tcpVersion int32
	state      int32
}

func (c *Conn) Error(tcpVersion int32) {
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
	c.state = ConnStateErr
}

func (c *Conn) SetState(state int32) {
	c.state = state
}

func (c *Conn) GetState() int32 {
	return c.state
}

func (c *Conn) GetTcp() (*net.TCPConn, int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn, c.tcpVersion
}

func (c *Conn) SetTcp(tcpConn *net.TCPConn) {
	c.Close()
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.tcpConn = tcpConn
	c.tcpVersion++
	c.state = ConnStateStart
}

func (c *Conn) ClearTcp() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.tcpConn = nil
	c.tcpVersion++
	c.state = ConnStateInit
}

func (c *Conn) Close() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
}
