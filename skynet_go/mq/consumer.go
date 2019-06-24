package mq

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc/rpcU"
	"time"
)

type ConsumerI interface {
	DoMqMsg(*MqMsg)
	GetConn() ConnI
	GetDescribe() string
	Error(ConnI)
}

func NewConsumer(cI ConsumerI) *Consumer {
	c := &Consumer{}
	c.cI = cI
	rpcU.NewServer(c)
	c.mqConn = NewMqConn()
	return c
}

type Consumer struct {
	rpcU.ServerB
	cI     ConsumerI
	mqConn *MqConn
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
	if c.mqConn.IsOk() {
		return true
	}

	connI := c.cI.GetConn()
	if connI == nil {
		return false
	}

	c.mqConn.SetConn(connI)

	return true
}

func (c *Consumer) GetDescribe() string {
	return c.cI.GetDescribe()
}

func (c *Consumer) Error() {
	c.cI.Error(c.mqConn.GetConn())
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
