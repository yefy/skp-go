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
	rpcU.NewServer(c)
	c.cI = cI
	c.mqConn = NewMqConn()
	c.setReadTimeout(3)
	return c
}

type Consumer struct {
	rpcU.ServerB
	cI          ConsumerI
	mqConn      *MqConn
	millisecond int32
	ReadTimeout time.Duration
}

func (c *Consumer) setReadTimeout(timeout time.Duration) {
	c.ReadTimeout = timeout
}

func (c *Consumer) RPC_Describe() string {
	return c.cI.GetDescribe()
}

func (c *Consumer) GetDescribe() string {
	return c.cI.GetDescribe()
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
	connI := c.cI.GetConn()
	if connI == nil {
		return false
	}

	c.mqConn.SetConn(connI)

	return true
}

func (c *Consumer) Error() {
	c.cI.Error(c.mqConn.GetConn())
}

func (c *Consumer) initTimeOut() {
	c.millisecond = 0
}

func (c *Consumer) addConnTimeOut(millisecond int32) {
	c.millisecond += millisecond
	if c.millisecond > 10000 {
		log.Debug(c.GetDescribe() + " : GetConn timeout")
		c.millisecond = 0
	}
}

func (c *Consumer) OnReadMqMsg() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Debug(c.GetDescribe() + " : OnReadMqMsg stop")
			return
		}

		if !c.GetConn() {
			time.Sleep(100 * time.Millisecond)
			c.addConnTimeOut(100)
			continue
		}

		mqMsg, err := c.mqConn.ReadMqMsg(c.ReadTimeout)
		c.ReadTimeout = 0
		if err != nil {
			errCode := errorCode.GetCode(err)
			if errCode == errorCode.TimeOut {
				c.cI.DoMqMsg(nil)
			} else {
				c.Error()
				continue
			}
		}

		if mqMsg == nil {
			continue
		}

		c.cI.DoMqMsg(mqMsg)
	}
}
