package client

import (
	"bytes"
	"encoding/binary"
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpc"

	"github.com/golang/protobuf/proto"
)

func NewCHProducer(c *Client) *CHProducer {
	p := &CHProducer{}
	rpc.NewServer(p)
	p.c = c
	p.GetTcpConn()
	return p
}

type CHProducer struct {
	rpc.ServerBase
	c       *Client
	tcpConn *net.TCPConn
}

func (p *CHProducer) GetTcpConn() {
	if p.tcpConn == nil {
		p.tcpConn = p.c.GetTcpConn()
	}
}

func (p *CHProducer) SendMqMsg(m *mq.MqMsg) {
	log.Fatal("msg = %+v", proto.MarshalTextString(m))
	mb, err := proto.Marshal(m)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	// if false {
	// 	mqMsg2 := MqMsg{}
	// 	proto.Unmarshal(mqMsgByte, &mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", proto.MarshalTextString(&mqMsg2))

	// }

	p.Write(mb)

}

func (p *CHProducer) WriteSize(size int) {
	log.Fatal("size = %d", size)
	s := int32(size)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, s)
	p.tcpConn.Write(bytesBuffer.Bytes())
}

func (p *CHProducer) Write(b []byte) {
	p.WriteSize(len(b))
	p.tcpConn.Write(b)
}
