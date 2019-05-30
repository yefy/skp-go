package rpc

import (
	"container/list"
	log "skp-go/skynet_go/logger"
	"sync"
	"testing"
)

//go test -test.bench=. server_cond_benchmark_test.go

type CondTest struct {
	sendMutex  *sync.Mutex
	sendCond   *sync.Cond
	sendList   *list.List
	recvMutex  *sync.Mutex
	recvCond   *sync.Cond
	recvList   *list.List
	done       chan interface{}
	typ        string
	number     int
	sendNumber int
	recvNumber int
}

func NewCondTest() *CondTest {
	c := &CondTest{}
	c.sendMutex = new(sync.Mutex)
	c.sendCond = sync.NewCond(c.sendMutex)
	c.sendList = list.New()
	c.recvMutex = new(sync.Mutex)
	c.recvCond = sync.NewCond(c.recvMutex)
	c.recvList = list.New()
	c.done = make(chan interface{}, 1)
	c.typ = ""
	c.number = 0
	c.sendNumber = 0
	c.recvNumber = 0
	return c
}

func (c *CondTest) Recv() {
	var sendData interface{} = 0
	for {
		c.sendCond.L.Lock()
		for {
			frontData := c.sendList.Front()
			if frontData != nil {
				c.sendList.Remove(frontData)
				sendData = frontData.Value
				break
			}
			c.sendCond.Wait()
		}
		//log.Debug("sendData = %+v", sendData.(int))
		c.sendCond.L.Unlock()
		if c.typ == "call" {
			c.recvCond.L.Lock()
			c.recvList.PushBack(sendData)
			c.recvCond.Signal()
			c.recvCond.L.Unlock()
		}
		c.recvNumber++
		if c.recvNumber == c.sendNumber {
			c.done <- 1
			return
		}
	}
}

func (c *CondTest) Start() {
	go c.Recv()
}

func (c *CondTest) Close() {
	<-c.done
	//log.Fatal("Close")
}

func (c *CondTest) Call() error {
	c.number++
	sendData := c.number
	c.sendCond.L.Lock()
	c.sendList.PushBack(sendData)
	c.sendCond.Signal()
	c.sendCond.L.Unlock()

	var recvData interface{} = 0
	c.recvCond.L.Lock()
	for {
		frontData := c.recvList.Front()
		if frontData != nil {
			c.recvList.Remove(frontData)
			recvData = frontData.Value
			break
		}
		c.recvCond.Wait()
	}
	c.recvCond.L.Unlock()
	if sendData != recvData.(int) {
		panic("Call error")
	}
	//log.Debug("recvData = %+v", recvData.(int))
	return nil
}

func (c *CondTest) Send() error {
	c.number++
	c.sendCond.L.Lock()
	c.sendList.PushBack(c.number)
	c.sendCond.Signal()
	c.sendCond.L.Unlock()
	return nil
}

func Benchmark_cond_Call(b *testing.B) {
	log.SetLevel(log.Lnone)
	c := NewCondTest()
	c.typ = "call"
	c.sendNumber = b.N
	c.Start()

	for i := 0; i < b.N; i++ {
		c.Call()
	}
	c.Close()
}

func Benchmark_cond_Send(b *testing.B) {
	log.SetLevel(log.Lnone)
	c := NewCondTest()
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
// Benchmark_rpc_cond_Call-4        1000000              1359 ns/op
// Benchmark_rpc_cond_Send-4       10000000               159 ns/op
// PASS
// ok      command-line-arguments  3.265s
