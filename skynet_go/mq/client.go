package mq

import (
	"reflect"
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
)

type ClientI interface {
	GetDescribe() string
	Error(ConnI)
}

func NewClient(clientI ClientI, connI ConnI) *Client {
	c := &Client{}
	c.clientI = clientI
	rpcU.NewServer(c)
	c.SetConn(connI)

	return c
}

type Client struct {
	rpcU.ServerB
	clientI ClientI
	mutex   sync.Mutex
	connI   ConnI
	state   int32
}

func (c *Client) RPC_Describe() string {
	return c.clientI.GetDescribe()
}

func (c *Client) GetDescribe() string {
	return c.clientI.GetDescribe()
}

func (c *Client) IsError(connI ConnI) bool {
	if (c.GetState() & (ClientStateStopping | ClientStateStop)) > 0 {
		return false
	}

	if (c.GetState() & ClientStateErr) > 0 {
		return false
	}

	if reflect.ValueOf(c.connI).Pointer() != reflect.ValueOf(connI).Pointer() {
		return false
	}

	c.DelState(ClientStateStart)
	c.AddState(ClientStateErr)

	return true
}

func (c *Client) SetState(state int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.state = state
}

func (c *Client) AddState(state int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.state |= state
}

func (c *Client) DelState(state int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.state &= ^state
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
}

func (c *Client) ClearConn() ConnI {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	connI := c.connI
	c.connI = nil

	return connI
}

func (c *Client) CloseConn() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.closeConn()
}

func (c *Client) closeConn() {
	if c.connI != nil {
		c.connI.Close()
		c.connI = nil
	}
}
