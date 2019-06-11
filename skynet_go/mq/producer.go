package mq

// import (
// 	"net"
// 	"skp-go/skynet_go/errorCode"
// 	log "skp-go/skynet_go/logger"
// 	"skp-go/skynet_go/rpc/rpc"
// 	"skp-go/skynet_go/utility"
// 	"time"

// 	"github.com/golang/protobuf/proto"
// )

// func NewProducer(conn *Conn) *Producer {
// 	p := &Producer{}
// 	p.conn = conn
// 	p.ProducerInterface = p
// 	rpc.NewServer(p)
// 	return p
// }

// type ProducerInterface interface {
// 	GetTcpConn() bool
// 	Error()
// }

// type Producer struct {
// 	rpc.ServerBase
// 	ProducerInterface
// 	conn       *Conn
// 	tcpConn    *net.TCPConn
// 	tcpVersion int32
// }

// func (p *Producer) Start() {
// 	p.Stop()
// 	p.RPC_GetServer().Start(false)
// }

// func (p *Producer) Stop() {
// 	p.RPC_GetServer().Stop(false)
// }

// func (p *Producer) GetTcpConn() bool {
// 	if p.tcpConn == nil && (p.conn.GetState()&connStateStart) > 0 {
// 		tcpConn, tcpVersion := p.conn.GetTcpConn()
// 		p.tcpVersion = tcpVersion
// 		p.tcpConn = tcpConn
// 	}

// 	if p.tcpConn != nil {
// 		return true
// 	}
// 	return false
// }

// func (p *Producer) Error() {
// 	p.conn.SendError(p.tcpVersion)
// 	p.tcpConn = nil
// }

// func (p *Producer) SendWriteMqMsg(mqMsg *MqMsg) {
// 	p.RPC_GetServer().Send("OnWriteMqMsg", mqMsg)
// }

// func (p *Producer) OnWriteMqMsg(mqMsg *MqMsg) error {
// 	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
// 	mqMsgBytes, err := proto.Marshal(mqMsg)
// 	if err != nil {
// 		return log.Panic(errorCode.NewErrCode(0, err.Error()))
// 	}
// 	for {
// 		if p.RPC_GetServer().IsStop() {
// 			log.Fatal("rpcRead stop")
// 			return nil
// 		}

// 		if !p.ProducerInterface.GetTcpConn() {
// 			time.Sleep(100 * time.Millisecond)
// 			continue
// 		}

// 		if err := p.Write(mqMsgBytes); err != nil {
// 			p.ProducerInterface.Error()
// 			continue
// 		}
// 		break
// 	}

// 	return nil
// }

// func (p *Producer) WriteMqMsg(mqMsg *MqMsg) error {
// 	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
// 	mqMsgBytes, err := proto.Marshal(mqMsg)
// 	if err != nil {
// 		return log.Panic(errorCode.NewErrCode(0, err.Error()))
// 	}

// 	if !p.ProducerInterface.GetTcpConn() {
// 		return log.Panic(errorCode.NewErrCode(0, "not GetTcpConn"))
// 	}

// 	if err := p.Write(mqMsgBytes); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p *Producer) Write(bytes []byte) error {
// 	if err := p.WriteSize(len(bytes)); err != nil {
// 		return err
// 	}

// 	err := p.WriteBytes(bytes)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p *Producer) WriteSize(size int) error {
// 	var err error
// 	bytes, err := utility.IntToBytes(size)
// 	if err != nil {
// 		return err
// 	}

// 	err = p.WriteBytes(bytes)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p *Producer) WriteBytes(bytes []byte) error {
// 	size := len(bytes)
// 	for size > 0 {
// 		wSize, err := p.tcpConn.Write(bytes)
// 		if err != nil {
// 			log.Fatal(err.Error())
// 			return err
// 		}

// 		if wSize == size {
// 			break
// 		}
// 		bytes = bytes[wSize:]
// 		size -= wSize
// 	}
// 	return nil
// }
