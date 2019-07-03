package rpc

import (
	"context"
	"reflect"
	log "skp-go/skynet_go/logger"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TypCall int32 = iota
	TypCallReq
	TypSend
	TypSendReq
)

const (
	stateStop = iota
	stateStart
	stateStopping
)

type CallBack1 func(error)
type CallBack2 func(string, error)

type Msg struct {
	Typ     int32
	Method  string
	Args    interface{}
	Reply   interface{}
	Pending chan interface{}
	Err     interface{}
	CB1     CallBack1
	CB2     CallBack2
	Encode  int32
}

func (m *Msg) Init() {
	m.Typ = 0
	m.Method = ""
	m.Err = nil
	m.Encode = 0
}

type ServerI interface {
	RPC_Describe() string
	RPC_Start()
	RPC_Stop()
	RPC_DoMsg(msg *Msg)
}

type Server struct {
	sI              ServerI
	cacheNumber     int
	Cache           chan interface{}
	goroutineNumber int32
	waitGroup       *sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	MsgPool         *sync.Pool
	ValuePools      []*sync.Pool
	SendNumber      int32
	recvNumber      int32
	state           int32
	waitMsg         int32
	mutex           sync.Mutex
}

func NewServer(sI ServerI) *Server {
	server := &Server{}
	server.sI = sI
	server.cacheNumber = 1000
	server.state = stateStop

	server.MsgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.Pending = make(chan interface{}, 1)
		return msg
	},
	}

	server.ValuePools = make([]*sync.Pool, 8)
	for i := 0; i < len(server.ValuePools); i++ {
		valueLen := i + 1
		server.ValuePools[i] = &sync.Pool{New: func() interface{} {
			value := make([]reflect.Value, valueLen)
			return value
		},
		}
	}

	server.sI.RPC_Start()

	server.Start(true)

	return server
}

func (server *Server) Addoroutine(num int) {
	for i := 0; i < num; i++ {
		atomic.AddInt32(&server.goroutineNumber, 1)
		server.waitGroup.Add(1)
		goroutineNumber := server.goroutineNumber
		go server.run(goroutineNumber)
	}
}

func (server *Server) Start(isNewCache bool) {
	defer server.mutex.Unlock()
	server.mutex.Lock()

	if server.state == stateStop {
		server.goroutineNumber = 0
		server.SendNumber = 0
		server.recvNumber = 0
		server.state = stateStart
		server.waitMsg = 1
		if isNewCache {
			server.Cache = make(chan interface{}, server.cacheNumber)
		}
		if server.Cache == nil {
			server.Cache = make(chan interface{}, server.cacheNumber)
		}
		server.waitGroup = &sync.WaitGroup{}
		server.ctx, server.cancel = context.WithCancel(context.Background())
		server.Addoroutine(1)
	}
}

func (server *Server) StopSelf(waitMsg bool) {
	go server.Stop(waitMsg)

}

func (server *Server) Stop(waitMsg bool) {
	defer server.mutex.Unlock()
	server.mutex.Lock()

	if server.state == stateStart {
		server.state = stateStopping
		if waitMsg {
			atomic.StoreInt32(&server.waitMsg, 1)
		} else {
			atomic.StoreInt32(&server.waitMsg, 0)
		}
		server.cancel()
		server.waitGroup.Wait()
		server.state = stateStop
		server.sI.RPC_Stop()
	}
}

func (server *Server) IsStart() bool {
	return atomic.LoadInt32(&server.state) == stateStart
}

func (server *Server) IsStop() bool {
	return atomic.LoadInt32(&server.state) == stateStopping
}

func (server *Server) isStopping() bool {
	if atomic.LoadInt32(&server.state) == stateStopping {
		if atomic.LoadInt32(&server.waitMsg) == 0 {
			return true
		} else {
			if atomic.LoadInt32(&server.recvNumber) == atomic.LoadInt32(&server.SendNumber) {
				return true
			}
		}
	}
	return false
}

func (server *Server) run(index int32) {
	defer server.waitGroup.Done()
	done := server.ctx.Done()
	//timer := time.NewTimer(time.Second)
	//timer.Reset(time.Second)
	//一次定时器  需要timer.Reset重新开始
	//timer := time.NewTicker(time.Second)
	//重复定时器

	var timer *time.Ticker
	var timerC <-chan time.Time = nil
	for {
		select {
		case <-timerC:
			if server.isStopping() {
				return
			}
			log.Fatal("RPC_Describe = %s, server.SendNumber = %d, server.recvNumber = %d", server.sI.RPC_Describe(), server.SendNumber, server.recvNumber)
		case <-done:
			done = nil
			if server.isStopping() {
				return
			}
			timer = time.NewTicker(time.Second)
			timerC = timer.C
		case cache := <-server.Cache:
			msg := cache.(*Msg)
			server.sI.RPC_DoMsg(msg)
			atomic.AddInt32(&server.recvNumber, 1)
			if server.isStopping() {
				return
			}
		}
	}
}

func (server *Server) RunOnce(callBack WaitCallBack) {
	server.waitGroup.Add(1)
	go func() {
		defer server.waitGroup.Done()
		callBack()
	}()
}

func (server *Server) Ticker(t time.Duration, callBack WaitCallBack2) {
	server.RunOnce(func() {
		timer := time.NewTicker(t)
		for {
			select {
			case <-timer.C:
				if callBack() {
					return
				}
			}

			if server.IsStop() {
				return
			}
		}
	})
}

func (server *Server) Timer(t time.Duration, callBack WaitCallBack) bool {
	timer := time.NewTimer(t)
	ch := make(chan bool)
	go func() {
		callBack()
		ch <- true
	}()
	select {
	case <-timer.C:
		return true
	case <-ch:
		return false
	}
}
