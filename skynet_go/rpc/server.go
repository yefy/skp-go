package rpc

import (
	"context"
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"sync"
	"sync/atomic"
)

var isTest bool = true

func SetTest(b bool) {
	isTest = b
}

const (
	typCall    = "Call"
	typCallReq = "CallReq"
	typSend    = "Send"
	typSendReq = "SendReq"
)

const (
	stateStop = iota
	stateStart
	stateStopping
)

type Msg struct {
	typ     string
	method  string
	args    interface{}
	reply   interface{}
	pending chan interface{}
	err     interface{}
}

func (m *Msg) init() {
	m.typ = ""
	m.method = ""
	m.err = nil
}

type ObjMethod struct {
	name      string
	method    reflect.Method
	typ       reflect.Type
	value     reflect.Value
	argvIn    int
	argvOut   int
	argType   reflect.Type
	replyType reflect.Type
	argName   string
	replyName string
}

type Service struct {
	obj       ServerInterface
	objName   string
	objType   reflect.Type
	objValue  reflect.Value
	objKind   reflect.Kind
	objMethod map[string]*ObjMethod
}

func NewService(obj ServerInterface) (*Service, error) {
	s := &Service{}
	s.obj = obj
	s.objType = reflect.TypeOf(obj)
	s.objKind = s.objType.Kind()
	if s.objKind != reflect.Ptr {
		return nil, log.Panic(errorCode.NewErrCode(0, "server.objKind != reflect.Ptr"))
	}
	s.objValue = reflect.ValueOf(obj)
	s.objName = reflect.Indirect(s.objValue).Type().Name()

	s.objMethod = make(map[string]*ObjMethod)
	for i := 0; i < s.objType.NumMethod(); i++ {
		method := s.objType.Method(i)

		objM := &ObjMethod{}
		objM.method = method
		objM.typ = method.Type
		objM.name = method.Name
		objM.value = s.objValue.MethodByName(method.Name)
		objM.argvIn = objM.value.Type().NumIn()
		objM.argvOut = objM.value.Type().NumOut()

		s.objMethod[objM.name] = objM
	}

	return s, nil
}

type ServerInterface interface {
	RPC_Server(*Server)
	RPC_Close()
}

type Server struct {
	service         *Service
	cacheNumber     int
	cache           chan interface{}
	goroutineNumber int32
	waitGroup       *sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	msgPool         *sync.Pool
	valuePools      []*sync.Pool
	sendNumber      int32
	recvNumber      int32
	state           int32
	waitMsg         int32
}

func NewServer(obj ServerInterface) *Server {
	service, err := NewService(obj)
	if err != nil {
		return nil
	}

	server := &Server{}
	server.service = service
	server.cacheNumber = 1000
	server.state = stateStop

	server.msgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}

	server.valuePools = make([]*sync.Pool, 8)
	for i := 0; i < len(server.valuePools); i++ {
		valueLen := i + 1
		server.valuePools[i] = &sync.Pool{New: func() interface{} {
			value := make([]reflect.Value, valueLen)
			return value
		},
		}
	}

	obj.RPC_Server(server)

	server.Start()

	return server
}

func (server *Server) object() interface{} {
	return server.service.obj
}

func (server *Server) Addoroutine(num int) {
	for i := 0; i < num; i++ {
		atomic.AddInt32(&server.goroutineNumber, 1)
		server.waitGroup.Add(1)
		goroutineNumber := server.goroutineNumber
		go server.run(goroutineNumber)
	}
}

func (server *Server) Start() {
	if server.state == stateStop {
		server.goroutineNumber = 0
		server.sendNumber = 0
		server.recvNumber = 0
		server.state = stateStart
		server.waitMsg = 1
		server.cache = make(chan interface{}, server.cacheNumber)
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
		server.service.obj.RPC_Close()
	}
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
	objMethod := service.objMethod[msg.method]
	args := msg.args.([]interface{})

	if msg.typ == typSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()
			valuePool := server.valuePools[len(args)]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args); i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			objMethod.method.Func.Call(argsValue)
		}()
		server.msgPool.Put(msg)
	} else if msg.typ == typSendReq {
		var replyValues []reflect.Value
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()
			valuePool := server.valuePools[len(args)-1]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args)-1; i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			replyValues = objMethod.method.Func.Call(argsValue)

			replyFunc := args[len(args)-1]
			replyFuncName := reflect.ValueOf(replyFunc)
			replyFuncName.Call(replyValues)
		}()

		server.msgPool.Put(msg)
	} else if msg.typ == typCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.err = log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			valuePool := server.valuePools[len(args)]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args); i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			replyValues := objMethod.method.Func.Call(argsValue)
			msg.reply = replyValues
		}()
		msg.pending <- msg
	} else if msg.typ == typCallReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.err = log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			valuePool := server.valuePools[len(args)-1]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args)-1; i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			replyValues := objMethod.method.Func.Call(argsValue)
			msg.reply = replyValues
		}()
		msg.pending <- msg
	}
	//objMethod.value.Call([]reflect.Value{reflect.ValueOf(args)})
	//objMethod.method.Func.Call([]reflect.Value{objValue, reflect.ValueOf(args)})
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

func checkArgv(isFunc bool, args ...interface{}) error {
	if isTest == false {
		return nil
	}

	for i := 0; i < len(args); i++ {
		argsKind := reflect.ValueOf(args[i]).Kind()
		if argsKind == reflect.Invalid {
			return log.Panic(errorCode.NewErrCode(0, "args index = %+v Invalid", i))
		}
		if isFunc && i == len(args)-1 && argsKind != reflect.Func {
			return log.Panic(errorCode.NewErrCode(0, "last not func"))
		}
	}

	return nil
}

func (server *Server) Send(method string, args ...interface{}) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	if err := checkArgv(false, args...); err != nil {
		return err
	}

	msg := server.msgPool.Get().(*Msg)
	msg.init()
	msg.typ = typSend
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	return nil
}

func (server *Server) SendReq(method string, args ...interface{}) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	if err := checkArgv(true, args...); err != nil {
		return err
	}

	msg := server.msgPool.Get().(*Msg)
	msg.init()
	msg.typ = typSendReq
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	return nil
}

func (server *Server) Call(method string, args ...interface{}) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	if err := checkArgv(false, args...); err != nil {
		return err
	}

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
		return msg.err.(error)
	}

	replyValues := msg.reply.([]reflect.Value)
	if len(replyValues) > 0 {
		first, ok := replyValues[0].Interface().(error)
		if ok {
			return first
		}
	}

	return nil
}

func (server *Server) CallReq(method string, args ...interface{}) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	if err := checkArgv(true, args...); err != nil {
		return err
	}

	msg := server.msgPool.Get().(*Msg)
	msg.init()
	defer server.msgPool.Put(msg)
	msg.typ = typCallReq
	msg.method = method
	msg.args = args

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	<-msg.pending

	if msg.err == nil {
		replyFunc := args[len(args)-1]
		replyFuncName := reflect.ValueOf(replyFunc)
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.err = log.Panic(errorCode.NewErrCode(0, "%+v", err))
				}
			}()
			replyFuncName.Call(msg.reply.([]reflect.Value))
		}()
	}

	if msg.err != nil {
		return msg.err.(error)
	}

	return nil
}
