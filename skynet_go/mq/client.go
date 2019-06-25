package mq

import (
	"reflect"
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
)

func NewClient(connI ConnI) *Client {
	c := &Client{}
	rpcU.NewServer(c)
	c.SetConn(connI)

	return c
}

type Client struct {
	rpcU.ServerB
	mutex sync.Mutex
	connI ConnI
	state int32
}

func (c *Client) Error(connI ConnI) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.state == ClientStateErr {
		return
	}

	if reflect.ValueOf(c.connI).Pointer() != reflect.ValueOf(connI).Pointer() {
		return
	}

	c.closeConn()
	c.state = ClientStateErr
}

func (c *Client) SetState(state int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.state = state
}

func (c *Client) GetState() int32 {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	return c.state
}

func (c *Client) GetConn() ConnI {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	return c.connI
}

func (c *Client) SetConn(connI ConnI) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.closeConn()

	c.connI = connI
	c.state = ClientStateStart
}

func (c *Client) ClearConn2() ConnI {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	connI := c.connI
	c.connI = nil
	c.state = ClientStateInit

	return connI
}

func (c *Client) CloseConn() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.closeConn()
	c.state = ClientStateInit
}

func (c *Client) closeConn() {
	if c.connI != nil {
		c.connI.Close()
		c.connI = nil
	}
}
