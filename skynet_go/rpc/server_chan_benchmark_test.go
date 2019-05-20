package rpc

import (
	log "skp-go/skynet_go/logger"
	"testing"
)

//go test -test.bench=. server_chan_benchmark_test.go

type ChanTest struct {
	send       chan interface{}
	recv       chan interface{}
	done       chan interface{}
	typ        string
	number     int
	sendNumber int
	recvNumber int
}

func NewChanTest(sendChanNumber int) *ChanTest {
	c := &ChanTest{}
	c.send = make(chan interface{}, sendChanNumber)
	c.recv = make(chan interface{}, 1)
	c.done = make(chan interface{}, 1)
	c.typ = ""
	c.number = 0
	c.sendNumber = 0
	c.recvNumber = 0
	return c
}

func (c *ChanTest) Recv() {
	for {
		sendData := <-c.send
		//log.Debug("sendData = %+v", sendData.(int))
		if c.typ == "call" {
			c.recv <- sendData
		}
		c.recvNumber++
		if c.recvNumber == c.sendNumber {
			c.done <- 1
			return
		}
	}
}

func (c *ChanTest) Start() {
	go c.Recv()
}

func (c *ChanTest) Close() {
	<-c.done
	//log.Fatal("Close")
}

func (c *ChanTest) Call() error {
	c.number++
	sendData := c.number
	c.send <- sendData
	reveData := <-c.recv
	if sendData != reveData.(int) {
		panic("Call error")
	}
	//log.Debug("reveData = %+v", reveData.(int))
	return nil
}

func (c *ChanTest) Send() error {
	c.number++
	c.send <- c.number
	return nil
}

func Benchmark_chan_Call(b *testing.B) {
	log.SetLevel(log.Lnone)
	c := NewChanTest(1)
	c.typ = "call"
	c.sendNumber = b.N
	c.Start()
	for i := 0; i < b.N; i++ {
		c.Call()
	}
	c.Close()
}

func Benchmark_chan_Send(b *testing.B) {
	log.SetLevel(log.Lnone)
	c := NewChanTest(1000)
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
