package client

import (
	"bytes"
	"encoding/binary"
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewCHConsumer(client *Client) *CHConsumer {
	c := &CHConsumer{}
	rpc.NewServer(c)
	c.c = client
	c.tcpConn = c.c.GetTcpConn()
	c.vector = mq.NewVector()
	c.vector.SetConn(c.tcpConn)
	c.Start()
	return c
}

type CHConsumer struct {
	rpc.ServerBase
	c       *Client
	tcpConn *net.TCPConn
	vector  *mq.Vector
}

func (c *CHConsumer) Start() {
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("Read")
}

func (c *CHConsumer) Read() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		rMqMsg, err := c.ReadMqMsg(0)
		if err != nil {
			return
		}

		if rMqMsg.GetTyp() == mq.TypeRespond {
			pendingMsgI, ok := c.c.pendingMap.Load(rMqMsg.GetPendingSeq())
			if !ok {
				log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
				continue
			}
			c.c.pendingMap.Delete(rMqMsg.GetPendingSeq())
			pendingMsg := pendingMsgI.(*PendingMsg)
			if pendingMsg.typ == mq.TypeCall {
				pendingMsg.pending <- rMqMsg
			}
		} else {
			c.c.rpcEncode.SendReq(rMqMsg.GetEncode(), rMqMsg.GetMethod(), rMqMsg.GetBody(), func(outStr string, err error) {
				log.Fatal("outStr = %+v, err = %+v", outStr, err)
				sMqMsg, err := mq.ReplyMqMsg(rMqMsg.GetHarbor(), rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), outStr)
				if err != nil {
					log.Err(err.Error())
					return
				}
				c.c.chProducer.SendMqMsg(sMqMsg)
			})
		}
	}
}

func (c *CHConsumer) ReadMqMsg(timeout time.Duration) (*mq.MqMsg, error) {
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

		//获取msg 如果有返回  如果没有接收数据 超时返回错误
		log.Fatal("c.vector.read(timeout) timeout = %d", timeout)
		if err := c.vector.Read(timeout); err != nil {
			return nil, errorCode.NewErrCode(0, err.Error())
		}
	}

	return nil, nil
}

func (c *CHConsumer) getMqMsgSize() int {
	sizeByte := c.vector.Get(4)
	if sizeByte == nil {
		return 0
	}
	bytesBuffer := bytes.NewBuffer(sizeByte)
	var size int32
	if err := binary.Read(bytesBuffer, binary.BigEndian, &size); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
		return 0
	}

	if !c.vector.CheckSize(int(4 + size)) {
		return 0
	}

	c.vector.Skip(4)
	return int(size)
}

func (c *CHConsumer) getMqMsg() (*mq.MqMsg, error) {
	size := c.getMqMsgSize()
	log.Fatal("CHConsumer size = %d", size)
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

/*
func (c *CHConsumer) Read() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		rMqMsg, err := c.ReadMqMsg(0)
		if err != nil {
			continue
		}
		if rMqMsg.Ty

		sMqMsg := &MqMsg{}
	sMqMsg.Typ = proto.Int32(typRespond)
	sMqMsg.Harbor = proto.Int32(conn.harbor)
	sMqMsg.PendingSeq = proto.Uint64(rMqMsg.PendingSeq)
	sMqMsg.Encode = proto.Int32(encodeGob)
	sMqMsg.Body = proto.String("")
	conn.shConsumer.SendMqMsg(sMqMsg)

	}
}

func (c *CHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil, nil
		}
		//获取msg 如果有返回  如果没有接收数据 超时返回错误
		c.vector.read(timeout)
	}

	return nil, nil
}
*/
