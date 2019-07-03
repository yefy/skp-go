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
	rpcU.NewServer(p)
	p.pI = pI
	p.mqConn = NewMqConn()

	// p.RPC_GetServer().Timer(time.Second, func() bool {
	// 	p.OnTimeOut()
	// 	return false
	// })

	return p
}

type Producer struct {
	rpcU.ServerB
	pI     ProducerI
	mqConn *MqConn
}

func (p *Producer) RPC_GetDescribe() string {
	return "NewProducer"
}

func (p *Producer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *Producer) Stop() {
	p.RPC_GetServer().Stop(false)
}

func (p *Producer) GetConn() bool {
	// if p.mqConn.IsOk() {
	// 	return true
	// }

	connI := p.pI.GetConn()
	if connI == nil {
		return false
	}

	if p.mqConn.SetConn(connI) {
		//___yefy
		//重新发送 未响应的数据包
	}

	return true
}

func (p *Producer) GetDescribe() string {
	return p.pI.GetDescribe()
}

func (p *Producer) Error() {
	p.pI.Error(p.mqConn.GetConn())
}

// func (p *Producer) OnTimeOut() {
// 	log.Fatal("OnTimeOut")
// 	p.SendWriteMqMsg(nil)
// }

func (p *Producer) RpcSend_OnWriteMqMsg(mqMsg *MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", mqMsg)
}

func (p *Producer) OnWriteMqMsg(mqMsg *MqMsg) error {
	for {
		if p.RPC_GetServer().IsStop() {
			log.Fatal(p.GetDescribe() + " : Producer OnWriteMqMsg stop")
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
