package rpc

import (
	log "skp-go/skynet_go/logger"
	"sync"
	"testing"
)

//go test -test.bench=. server_chanEx_benchmark_test.go

type ChanTestEx struct {
	send       chan interface{}
	recv       chan interface{}
	done       chan interface{}
	typ        string
	number     int
	sendNumber int
	recvNumber int
	server     *ServerTest_b
	err        error
	msgPool    *sync.Pool
}

func NewChanTestEx(sendChanNumber int) *ChanTestEx {
	c := &ChanTestEx{}
	c.send = make(chan interface{}, sendChanNumber)
	c.recv = make(chan interface{}, 1)
	c.done = make(chan interface{}, 1)
	c.typ = ""
	c.number = 0
	c.sendNumber = 0
	c.recvNumber = 0
	c.server = NewServerTest_b()

	c.msgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}
	return c
}

func (c *ChanTestEx) Recv() {
	for {
		send := <-c.send
		msg := send.(*Msg)
		sendData := msg.args
		//log.Debug("sendData = %+v", sendData.(int))

		out := 0
		err := c.server.ExampleSuccess(sendData.(int), &out)
		if c.typ == "call" {
			msg.reply = out
			msg.err = err
			c.recv <- msg
		} else {
			c.msgPool.Put(msg)
		}
		c.recvNumber++
		if c.recvNumber == c.sendNumber {
			c.done <- 1
			return
		}
	}
}

func (c *ChanTestEx) Start() {
	go c.Recv()
}

func (c *ChanTestEx) Close() {
	<-c.done
	//log.Fatal("Close")
}

func (c *ChanTestEx) Call() error {
	c.number++
	sendData := c.number
	msg := c.msgPool.Get().(*Msg)
	msg.args = sendData
	c.send <- msg
	recv := <-c.recv
	reveMsg := recv.(*Msg)
	reveData := reveMsg.reply
	if sendData != reveData.(int) {
		panic("Call error")
	}
	//log.Debug("reveData = %+v", reveData.(int))
	return nil
}

func (c *ChanTestEx) Send() error {
	c.number++
	msg := c.msgPool.Get().(*Msg)
	msg.args = c.number
	c.send <- msg
	return nil
}

func Benchmark_ExampleSuccess_chanData_Call(b *testing.B) {
	b.ReportAllocs()
	log.SetLevel(log.Lnone)
	c := NewChanTestEx(1)
	c.typ = "call"
	c.sendNumber = b.N
	c.Start()
	for i := 0; i < b.N; i++ {
		c.Call()
	}
	c.Close()
}

func Benchmark_ExampleSuccess_chanData_Send(b *testing.B) {
	b.ReportAllocs()
	log.SetLevel(log.Lnone)
	c := NewChanTestEx(1000)
	c.typ = "send"
	c.sendNumber = b.N
	c.Start()
	for i := 0; i < b.N; i++ {
		c.Send()
	}
	c.Close()
}

// goos: windows
// goarch: amd64
// Benchmark_rpc_chan_Call-4        2000000               890 ns/op
// Benchmark_rpc_chan_Send-4       10000000               128 ns/op
// PASS
// ok      command-line-arguments  4.212s
