package server

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewSHConsumer(conn *Conn) *SHConsumer {
	c := &SHConsumer{}
	c.conn = conn
	rpc.NewServer(c)
	return c
}

type SHConsumer struct {
	rpc.ServerBase
	conn       *Conn
	vector     *mq.Vector
	tcpConn    *net.TCPConn
	tcpVersion int32
}

func (c *SHConsumer) Start() {
	c.Stop()
	c.RPC_GetServer().Start(false)
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("OnRead")
}

func (c *SHConsumer) Stop() {
	c.RPC_GetServer().Stop(true)
}

func (c *SHConsumer) GetTcpConn() bool {
	if c.tcpConn == nil && (c.conn.GetState()&mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := c.conn.GetTcpConn()
		if tcpVersion != c.tcpVersion {
			c.tcpVersion = tcpVersion
			c.vector = mq.NewVector()
		}
		c.tcpConn = tcpConn
		c.vector.SetConn(c.tcpConn)
	}

	if c.tcpConn != nil {
		return true
	}
	return false
}

func (c *SHConsumer) OnRead() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		if !c.GetTcpConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		mqMsg, err := c.ReadMqMsg(3)
		if err != nil {
			errCode := err.(*errorCode.ErrCode)
			if errCode.Code() != errorCode.TimeOut {
				c.tcpConn = nil
				c.conn.SendError(c.tcpVersion)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		if mqMsg == nil {
			continue
		}

		if mqMsg.GetTyp() == mq.TypeRespond {
			harbor := mqMsg.GetHarbor()
			harborConnI, ok := c.conn.server.harborConn.Load(harbor)
			if ok {
				harborConn := harborConnI.(*Conn)
				harborConn.shProducer.SendWriteMqMsg(mqMsg)
			} else {
				//这里需要保存mqMsg用于排查问题
				log.Fatal("not harbor = %d", harbor)
			}
		} else {
			key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
			q := c.conn.server.topicTag[key]
			if q == nil {
				log.Fatal("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
				q = NewSQProducer(c.conn.server, mqMsg.GetTopic(), mqMsg.GetTag())
				c.conn.server.topicTag[key] = q
			}
			q.SendWriteMqMsg(mqMsg)
		}
	}
}

func (c *SHConsumer) ReadMqMsg(timeout time.Duration) (*mq.MqMsg, error) {
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

func (c *SHConsumer) getMqMsgSize() (int, error) {
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

func (c *SHConsumer) getMqMsg() (*mq.MqMsg, error) {
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

	msg := &mq.MqMsg{}
	if err := proto.Unmarshal(msgByte, msg); err != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	c.vector.Skip(size)
	log.Fatal("msg = %+v", proto.MarshalTextString(msg))
	return msg, nil
}
