package mq

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc/rpcU"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

type ConsumerI interface {
	DoMqMsg(*MqMsg)
	GetConn() (conn.ConnI, int32, bool)
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
	vector      *Vector
	connI       conn.ConnI
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
	connI, connVersion, ok := c.cI.GetConn()
	if !ok {
		return ok
	}

	log.Fatal("connI = %+v, connVersion = %+v, c.connVersion = %+v", connI, connVersion, c.connVersion)
	if c.connVersion != connVersion {
		c.connI = connI
		c.connVersion = connVersion
		c.vector = NewVector()
		c.vector.SetConn(c.connI)
	} else {
		if c.connI == nil {
			c.connI = connI
		}
	}
	return ok
}

func (c *Consumer) GetDescribe() string {
	return c.cI.GetDescribe()
}

func (c *Consumer) Error() {
	c.cI.Error(c.connVersion)
}

func (c *Consumer) OnReadMqMsg() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		if !c.GetConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		mqMsg, err := c.ReadMqMsg(3)
		if err != nil {
			errCode := err.(*errorCode.ErrCode)
			if errCode.Code() != errorCode.TimeOut {
				c.connI = nil
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

func (c *Consumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil, nil
		}

		rMqMsg, err := c.getMqMsg()
		if err != nil {
			return nil, err
		}

		if rMqMsg != nil {
			return rMqMsg, nil
		}

		if err := c.vector.Read(timeout); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (c *Consumer) getMqMsgSize() (int, error) {
	sizeBytes := c.vector.Get(4)
	if sizeBytes == nil {
		return 0, nil
	}

	size, err := utility.BytesToInt(sizeBytes)
	if err != nil {
		return 0, err
	}

	if !c.vector.CheckSize(int(4 + size)) {
		return 0, nil
	}

	c.vector.Skip(4)
	return int(size), nil
}

func (c *Consumer) getMqMsg() (*MqMsg, error) {
	size, err := c.getMqMsgSize()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}
	msgByte := c.vector.Get(size)
	if msgByte == nil {
		return nil, nil
	}

	msg := &MqMsg{}
	if err := proto.Unmarshal(msgByte, msg); err != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	c.vector.Skip(size)
	log.Fatal("msg = %+v", proto.MarshalTextString(msg))
	return msg, nil
}
