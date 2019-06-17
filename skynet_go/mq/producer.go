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

type ProducerI interface {
	GetConn() (conn.ConnI, int32, bool)
	GetDescribe() string
	Error(int32)
}

func NewProducer(pI ProducerI) *Producer {
	p := &Producer{}
	p.pI = pI
	rpcU.NewServer(p)
	p.millisecond = 0
	return p
}

type Producer struct {
	rpcU.ServerB
	pI          ProducerI
	connI       conn.ConnI
	connVersion int32
	millisecond int32
}

func (p *Producer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *Producer) Stop() {
	p.RPC_GetServer().Stop(false)
}

func (p *Producer) GetConn() bool {
	connI, connVersion, ok := p.pI.GetConn()
	if !ok {
		return ok
	}

	if p.connVersion != connVersion {
		p.connVersion = connVersion
		p.connI = connI
		//___yefy
		//重新发送
	} else {
		if p.connI == nil {
			p.connI = connI
		}
	}
	return ok
}

func (p *Producer) GetDescribe() string {
	return p.pI.GetDescribe()
}

func (p *Producer) Error() {
	p.pI.Error(p.connVersion)
}

func (p *Producer) SendWriteMqMsg(mqMsg *MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", mqMsg)
}

func (p *Producer) OnWriteMqMsg(mqMsg *MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))

	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	for {
		if p.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil
		}

		if !p.GetConn() {
			time.Sleep(100 * time.Millisecond)
			p.millisecond += 100
			if p.millisecond > 1000*5 {
				p.millisecond = 0
				log.Err(p.pI.GetDescribe() + "  getconn timeout")
			}
			continue
		}

		if err := p.Write(mqMsgBytes); err != nil {
			p.Error()
			continue
		}
		break
	}

	return nil
}

func (p *Producer) Write(bytes []byte) error {
	if err := p.WriteSize(len(bytes)); err != nil {
		return err
	}

	err := p.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) WriteSize(size int) error {
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

func (p *Producer) WriteBytes(bytes []byte) error {
	size := len(bytes)
	for size > 0 {
		wSize, err := p.connI.Write(bytes)
		if err != nil {
			log.Fatal(err.Error())
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
