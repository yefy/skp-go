package server

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcdp"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewSHProducer(conn *Conn) *SHProducer {
	p := &SHProducer{}
	p.conn = conn
	p.SHProducerInterface = p
	rpcdp.NewServer(p)
	return p
}

type SHProducerInterface interface {
	GetTcpConn() bool
	Error()
}

type SHProducer struct {
	rpcdp.ServerBase
	SHProducerInterface
	conn       *Conn
	tcpConn    *net.TCPConn
	tcpVersion int32
}

func (p *SHProducer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *SHProducer) Stop() {
	p.RPC_GetServer().Stop(false)
}

func (p *SHProducer) GetTcpConn() bool {
	if p.tcpConn == nil && (p.conn.GetState()&mq.ConnStateStart) > 0 {
		tcpConn, tcpVersion := p.conn.GetTcpConn()
		p.tcpVersion = tcpVersion
		p.tcpConn = tcpConn
	}

	if p.tcpConn != nil {
		return true
	}
	return false
}

func (p *SHProducer) RPC_Dispath(method string, args []interface{}) error {
	if method == "OnWriteMqMsg" {
		mqMsg := args[0].(*mq.MqMsg)
		return p.OnWriteMqMsg(mqMsg)
	}

	return nil
}

func (p *SHProducer) Error() {
	p.conn.SendError(p.tcpVersion)
	p.tcpConn = nil
}

func (p *SHProducer) SendWriteMqMsg(m *mq.MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", m)
}

func (p *SHProducer) OnWriteMqMsg(mqMsg *mq.MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	for {
		if p.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil
		}

		if !p.SHProducerInterface.GetTcpConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := p.Write(mqMsgBytes); err != nil {
			p.SHProducerInterface.Error()
			continue
		}
		break
	}

	return nil
}

func (p *SHProducer) WriteMqMsg(mqMsg *mq.MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if !p.SHProducerInterface.GetTcpConn() {
		log.Panic(errorCode.NewErrCode(0, "not GetTcpConn"))
	}
	log.Fatal("mqMsgBytes size = %d", len(mqMsgBytes))
	if err := p.Write(mqMsgBytes); err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) WriteSize(size int) error {
	var err error
	bytes, err := utility.IntToBytes(size)
	if err != nil {
		return err
	}

	err = p.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) Write(bytes []byte) error {
	if err := p.WriteSize(len(bytes)); err != nil {
		return err
	}

	err := p.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) WriteBytes(bytes []byte) error {
	size := len(bytes)
	for size > 0 {
		wSize, err := p.tcpConn.Write(bytes)
		if err != nil {
			return err
		}

		if wSize == size {
			break
		}
		bytes = bytes[wSize:]
		size -= wSize
	}
	return nil
}
