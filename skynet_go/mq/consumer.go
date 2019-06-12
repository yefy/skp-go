package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc/rpc"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

type ConsumerI interface {
	DoMqMsg(*MqMsg)
	GetTcp() (*net.TCPConn, int32, bool)
	Error(int32)
}

func NewConsumer(cI ConsumerI) *Consumer {
	c := &Consumer{}
	c.cI = cI
	rpc.NewServer(c)
	return c
}

type Consumer struct {
	rpc.ServerBase
	cI         ConsumerI
	vector     *Vector
	tcpConn    *net.TCPConn
	tcpVersion int32
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

func (c *Consumer) GetTcp() bool {
	tcpConn, tcpVersion, ok := c.cI.GetTcp()
	if !ok {
		return ok
	}
	c.tcpConn = tcpConn
	log.Fatal("tcpConn = %+v, tcpVersion = %+v, c.tcpVersion = %+v", tcpConn, tcpVersion, c.tcpVersion)
	if c.tcpVersion != tcpVersion {
		c.tcpVersion = tcpVersion
		c.vector = NewVector()
	}
	c.vector.SetConn(c.tcpConn)
	return ok
}

func (c *Consumer) Error() {
	c.cI.Error(c.tcpVersion)
}

func (c *Consumer) OnReadMqMsg() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		if !c.GetTcp() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		mqMsg, err := c.ReadMqMsg(3)
		if err != nil {
			errCode := err.(*errorCode.ErrCode)
			if errCode.Code() != errorCode.TimeOut {
				c.tcpConn = nil
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
