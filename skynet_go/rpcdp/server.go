package rpcdp

import (
	"context"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"sync"
	"sync/atomic"
)

const (
	typCall    = "Call"
	typSend    = "Send"
	typSendReq = "SendReq"
)

const (
	stateStop = iota
	stateStart
	stateStopping
)

type Msg struct {
	typ      string
	method   string
	args     interface{}
	pending  chan interface{}
	callBack CallBack
	err      error
}

func (m *Msg) init() {
	m.typ = ""
	m.method = ""
	m.callBack = nil
	m.err = nil
}

type CallBack func(error)

type ServerInterface interface {
	RPC_SetServer(*Server)
	RPC_GetServer() *Server
	RPC_Dispath(method string, args []interface{}) error
	RPC_Close()
}

type ServerBase struct {
	server *Server
}

func (sb *ServerBase) RPC_SetServer(server *Server) {
	sb.server = server
}

func (sb *ServerBase) RPC_GetServer() *Server {
	return sb.server
}

func (sb *ServerBase) RPC_Close() {

}

type Server struct {
	service         ServerInterface
	cacheNumber     int
	cache           chan interface{}
	goroutineNumber int32
	waitGroup       *sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	msgPool         *sync.Pool
	sendNumber      int32
	recvNumber      int32
	state           int32
	waitMsg         int32
	mutex           sync.Mutex
}

func NewServer(obj ServerInterface) *Server {
	server := &Server{}
	server.service = obj
	server.cacheNumber = 1000
	server.state = stateStop

	server.msgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}

	obj.RPC_SetServer(server)

	server.Start(true)

	return server
}

func (server *Server) object() interface{} {
	return server.service
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
		server.sendNumber = 0
		server.recvNumber = 0
		server.state = stateStart
		server.waitMsg = 1
		if isNewCache {
			server.cache = make(chan interface{}, server.cacheNumber)
		}

		if server.cache == nil {
			server.cache = make(chan interface{}, server.cacheNumber)
		}
		server.waitGroup = &sync.WaitGroup{}
		server.ctx, server.cancel = context.WithCancel(context.Background())
		server.Addoroutine(1)
	}
}

func (server *Server) SendStop(waitMsg bool) {
	go func() {
		server.Stop(waitMsg)
	}()
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
		server.service.RPC_Close()
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
			if atomic.LoadInt32(&server.recvNumber) == atomic.LoadInt32(&server.sendNumber) {
				return true
			}
		}
	}
	return false
}

func (server *Server) callBack(msg *Msg) {
	service := server.service
	args := msg.args.([]interface{})

	if msg.typ == typSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			service.RPC_Dispath(msg.method, args)
		}()
		server.msgPool.Put(msg)
	} else if msg.typ == typSendReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			err := service.RPC_Dispath(msg.method, args)
			msg.callBack(err)
		}()

		server.msgPool.Put(msg)
	} else if msg.typ == typCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.err = log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			err := service.RPC_Dispath(msg.method, args)
			msg.err = err
		}()
		msg.pending <- msg
	}
}

func (server *Server) run(index int32) {
	defer server.waitGroup.Done()
	done := server.ctx.Done()
	for {
		select {
		case <-done:
			done = nil
			if server.isStopping() {
				return
			}
		case cache := <-server.cache:
			msg := cache.(*Msg)
			server.callBack(msg)
			atomic.AddInt32(&server.recvNumber, 1)
			if server.isStopping() {
				return
			}
		}
	}
}

func (server *Server) Send(method string, args ...interface{}) {
	msg := server.msgPool.Get().(*Msg)
	msg.init()
	msg.typ = typSend
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
}

func (server *Server) SendReq(callBack CallBack, method string, args ...interface{}) {
	msg := server.msgPool.Get().(*Msg)
	msg.init()
	msg.typ = typSendReq
	msg.callBack = callBack
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
}

func (server *Server) Call(method string, args ...interface{}) error {
	msg := server.msgPool.Get().(*Msg)
	msg.init()
	defer server.msgPool.Put(msg)
	msg.typ = typCall
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	<-msg.pending

	if msg.err != nil {
		return msg.err
	}

	return nil
}
