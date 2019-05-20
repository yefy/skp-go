package rpc

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"sync"
	"testing"
)

//go test -test.bench=. server_chanEx_benchmark_test.go

type ServiceTest_bEx struct {
	Num int
}

var errTestbEx error = errorCode.NewErrCode(0, "TestErr")

func NewServiceTest_bEx() *ServiceTest_bEx {
	serviceTest := &ServiceTest_bEx{}
	return serviceTest
}
func (serviceTest *ServiceTest_bEx) Test(in int, out *int) error {
	*out = in
	if serviceTest.Num != in {
		return log.Panic(errorCode.NewErrCode(0, "Test Num:%+v != in:%+v", serviceTest.Num, in))
	}
	log.Debug("Test in = %+v, out = %+v, Num = %+v \n", in, *out, serviceTest.Num)
	return nil
}

func (serviceTest *ServiceTest_bEx) TestErr(in int, out *int) error {
	*out = in
	if serviceTest.Num != in {
		return log.Panic(errorCode.NewErrCode(0, "TestErr Num:%+v != in:%+v", serviceTest.Num, in))
	}
	log.Debug("TestErr in = %+v, out = %+v, Num = %+v \n", in, *out, serviceTest.Num)
	return errTestbEx
}

type ChanTestEx struct {
	send       chan interface{}
	recv       chan interface{}
	done       chan interface{}
	typ        string
	number     int
	sendNumber int
	recvNumber int
	server     *ServiceTest_bEx
	err        error
	msgPool    *sync.Pool
}

type MsgEx struct {
	typ     string //call send
	method  string
	args    interface{}
	reply   interface{}
	pending chan interface{}
	err     interface{}
	//callValue []reflect.Value
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
	c.server = NewServiceTest_bEx()

	c.msgPool = &sync.Pool{New: func() interface{} {
		msg := &MsgEx{}
		msg.pending = make(chan interface{}, 1)
		//msg.callValue = make([]reflect.Value, 3)
		return msg
	},
	}
	return c
}

func (c *ChanTestEx) Recv() {
	for {
		msgint := <-c.send
		msg := msgint.(*MsgEx)
		sendData := msg.args
		log.Debug("sendData = %+v", sendData.(int))
		in := 1
		out := 0
		c.server.Num = in
		c.err = c.server.Test(in, &out)
		if c.typ == "call" {
			msg.reply = in
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
	log.Fatal("Close")
}

func (c *ChanTestEx) Call() error {
	c.number++
	msg := c.msgPool.Get().(*MsgEx)
	msg.args = c.number
	c.send <- msg
	revemsgint := <-c.recv
	revemsg := revemsgint.(*MsgEx)
	reveData := revemsg.reply
	log.Debug("reveData = %+v", reveData.(int))
	return nil
}

func (c *ChanTestEx) Send() error {
	c.number++
	msg := c.msgPool.Get().(*MsgEx)
	msg.args = c.number
	c.send <- msg
	return nil
}

func Benchmark_rpc_chanEx_Call(b *testing.B) {
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

func Benchmark_rpc_chanEx_Send(b *testing.B) {
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
