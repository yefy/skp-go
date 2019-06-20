package mq

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc/rpcU"
	"time"
)

type ConsumerI interface {
	DoMqMsg(*MqMsg)
	GetConn() (ConnI, int32, bool)
	GetDescribe() string
	Error(int32)
}

func NewConsumer(cI ConsumerI) *Consumer {
	c := &Consumer{}
	c.cI = cI
	rpcU.NewServer(c)
	return c
}

type Consumer struct {
	rpcU.ServerB
	cI          ConsumerI
	mqConn      *MqConn
	connVersion int32
}

func (c *Consumer) Start() {
	c.Stop()
	c.RPC_GetServer().Start(false)
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("OnReadMqMsg")
}

func (c *Consumer) Stop() {
	c.RPC_GetServer().Stop(true)
}

func (c *Consumer) GetConn() bool {
	if c.mqConn != nil {
		return true
	}

	connI, connVersion, ok := c.cI.GetConn()
	if !ok {
		return ok
	}

	if c.connVersion == connVersion {
		return false
	}

	c.connVersion = connVersion
	c.mqConn = NewMqConn()
	c.mqConn.SetConn(connI)

	return true
}

func (c *Consumer) GetDescribe() string {
	return c.cI.GetDescribe()
}

func (c *Consumer) Error() {
	c.mqConn = nil
	c.cI.Error(c.connVersion)
}

func (c *Consumer) OnReadMqMsg() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal(c.GetDescribe() + " : Consumer OnReadMqMsg stop")
			return
		}

		if !c.GetConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		mqMsg, err := c.mqConn.ReadMqMsg(3)
		if err != nil {
			errCode := errorCode.GetCode(err)
			if errCode == errorCode.TimeOut {
				c.cI.DoMqMsg(nil)
			} else {
				c.Error()
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		if mqMsg == nil {
			continue
		}

		c.cI.DoMqMsg(mqMsg)
	}
}
