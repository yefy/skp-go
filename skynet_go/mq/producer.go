package mq

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc/rpcU"
	"time"
)

type ProducerI interface {
	GetConn() ConnI
	GetDescribe() string
	Error(ConnI)
}

func NewProducer(pI ProducerI) *Producer {
	p := &Producer{}
	log.Fatal("NewProducer 000000000")
	rpcU.NewServer(p)
	log.Fatal("NewProducer 111111111")
	p.pI = pI
	p.mqConn = NewMqConn()
	log.Fatal("NewProducer 333333333")

	return p
}

type Producer struct {
	rpcU.ServerB
	pI          ProducerI
	mqConn      *MqConn
	millisecond int32
}

func (p *Producer) RPC_Describe() string {
	return p.pI.GetDescribe()
}

func (p *Producer) GetDescribe() string {
	return p.pI.GetDescribe()
}

func (p *Producer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *Producer) Stop() {
	p.RPC_GetServer().Stop(true)
}

func (p *Producer) GetConn() bool {
	connI := p.pI.GetConn()
	if connI == nil {
		return false
	}

	p.mqConn.SetConn(connI)

	return true
}

func (p *Producer) Error() {
	p.pI.Error(p.mqConn.GetConn())
}

func (p *Producer) initTimeOut() {
	p.millisecond = 0
}

func (p *Producer) addConnTimeOut(millisecond int32) {
	p.millisecond += millisecond
	if p.millisecond > 3000 {
		log.Debug(p.GetDescribe() + " : GetConn timeout")
		p.millisecond = 0
	}
}

func (p *Producer) RpcSend_OnWriteMqMsg(mqMsg *MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", mqMsg)
}

func (p *Producer) RpcCall_OnWriteMqMsg(mqMsg *MqMsg) error {
	return p.RPC_GetServer().Call("OnWriteMqMsg", mqMsg)
}

func (p *Producer) OnWriteMqMsg(mqMsg *MqMsg) error {
	p.initTimeOut()
	for {
		if p.RPC_GetServer().IsStop() {
			log.Debug(p.GetDescribe() + " :  OnWriteMqMsg stop")
			return nil
		}

		if !p.GetConn() {
			time.Sleep(100 * time.Millisecond)
			p.addConnTimeOut(100)
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
