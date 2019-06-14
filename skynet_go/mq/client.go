package mq

import (
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
)

func NewClient(tcpConn conn.ConnI) *Client {
	c := &Client{}
	rpcU.NewServer(c)
	if tcpConn != nil {
		c.SetTcp(tcpConn)
	}

	return c
}

type Client struct {
	rpcU.ServerB
	mutex      sync.Mutex
	tcpConn    conn.ConnI
	tcpVersion int32
	state      int32
}

func (c *Client) Error(tcpVersion int32) {
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
	c.state = ClientStateErr
}

func (c *Client) SetState(state int32) {
	c.state = state
}

func (c *Client) GetState() int32 {
	return c.state
}

func (c *Client) GetTcp() (conn.ConnI, int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn, c.tcpVersion
}

func (c *Client) SetTcp(tcpConn conn.ConnI) {
	c.Close()
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.tcpConn = tcpConn
	c.tcpVersion++
	c.state = ClientStateStart
}

func (c *Client) ClearTcp() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.tcpConn = nil
	c.tcpVersion++
	c.state = ClientStateInit
}

func (c *Client) Close() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
}
