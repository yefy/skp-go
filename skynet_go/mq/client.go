package mq

import (
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
)

func NewClient(connI ConnI) *Client {
	c := &Client{}
	rpcU.NewServer(c)
	if connI != nil {
		c.SetConn(connI)
	}

	return c
}

type Client struct {
	rpcU.ServerB
	mutex       sync.Mutex
	connI       ConnI
	connVersion int32
	state       int32
}

func (c *Client) Error(connVersion int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if connVersion != c.connVersion {
		return
	}
	if c.connI != nil {
		c.connI.Close()
		c.connI = nil
	}
	c.connVersion++
	c.state = ClientStateErr
}

func (c *Client) SetState(state int32) {
	c.state = state
}

func (c *Client) GetState() int32 {
	return c.state
}

func (c *Client) GetConn() (ConnI, int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.connI, c.connVersion
}

func (c *Client) SetConn(connI ConnI) {
	c.Close()
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.connI = connI
	c.connVersion++
	c.state = ClientStateStart
}

func (c *Client) ClearConn() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.connI = nil
	c.connVersion++
	c.state = ClientStateInit
}

func (c *Client) Close() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.connI != nil {
		c.connI.Close()
		c.connI = nil
	}
}
