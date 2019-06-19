package mq

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"skp-go/skynet_go/rpc/rpcU"
	"time"
)

type ProducerI interface {
	GetConn() (ConnI, int32, bool)
	GetDescribe() string
	Error(int32)
}

func NewProducer(pI ProducerI) *Producer {
	p := &Producer{}
	p.pI = pI
	rpcU.NewServer(p)
	//p.RPC_GetServer().Addoroutine(1)
	//p.RPC_GetServer().Send("OnTimeOut")
	return p
}

type Producer struct {
	rpcU.ServerB
	pI          ProducerI
	mqConn      *MqConn
	connVersion int32
}

func (p *Producer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *Producer) Stop() {
	p.RPC_GetServer().Stop(false)
}

func (p *Producer) GetConn() bool {
	if p.mqConn != nil {
		return true
	}

	connI, connVersion, ok := p.pI.GetConn()
	if !ok {
		return ok
	}

	if p.connVersion == connVersion {
		return false
	}

	p.connVersion = connVersion
	p.mqConn = NewMqConn()
	p.mqConn.SetConn(connI)
	//___yefy
	//重新发送 未响应的数据包

	return true
}

func (p *Producer) GetDescribe() string {
	return p.pI.GetDescribe()
}

func (p *Producer) Error() {
	p.mqConn = nil
	p.pI.Error(p.connVersion)
}

func (p *Producer) OnTimeOut() {
	rpc.Timer(time.Second, func() bool {
		log.Fatal("OnTimeOut")
		return false
	})
}

func (p *Producer) SendWriteMqMsg(mqMsg *MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", mqMsg)
}

func (p *Producer) OnWriteMqMsg(mqMsg *MqMsg) error {
	for {
		if p.RPC_GetServer().IsStop() {
			log.Fatal(p.GetDescribe() + " : Producer stop")
			return nil
		}

		if !p.GetConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := p.mqConn.WriteMqMsg(mqMsg); err != nil {
			p.Error()
			continue
		}
		break
	}

	return nil
}
