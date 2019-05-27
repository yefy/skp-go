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

type Server struct {
	obj         interface{}
	objName     string
	objType     reflect.Type
	objValue    reflect.Value
	objKind     reflect.Kind
	objMethod   map[string]*ObjMethod
	cacheNumber int
	cache       chan interface{}
	coNumber    int32
	waitGroup   *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	msgPool     *sync.Pool
	valuePools  []*sync.Pool
	sendNumber  int32
	recvNumber  int32
	done        int32
}

func parseMethod(server *Server, obj interface{}) error {
	server.objType = reflect.TypeOf(obj)
	server.objKind = server.objType.Kind()
	if server.objKind != reflect.Ptr {
		return log.Panic(errorCode.NewErrCode(0, "server.objKind != reflect.Ptr"))
	}
	server.objValue = reflect.ValueOf(obj)
	server.objName = reflect.Indirect(server.objValue).Type().Name()

	for i := 0; i < server.objType.NumMethod(); i++ {
		method := server.objType.Method(i)

		objM := &ObjMethod{}
		objM.method = method
		objM.typ = method.Type
		objM.name = method.Name
		objM.value = server.objValue.MethodByName(method.Name)
		objM.argvIn = objM.value.Type().NumIn()
		objM.argvOut = objM.value.Type().NumOut()

		server.objMethod[objM.name] = objM
	}
	return nil
}

func NewServer(obj interface{}) *Server {
	server := &Server{}
	server.obj = obj
	server.objMethod = make(map[string]*ObjMethod)
	server.cacheNumber = 1000
	server.cache = make(chan interface{}, server.cacheNumber)
	server.coNumber = 0
	server.sendNumber = 0
	server.recvNumber = 0
	server.done = 0

	err := parseMethod(server, obj)
	if err != nil {
		return nil
	}

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

	server.waitGroup = &sync.WaitGroup{}
	server.ctx, server.cancel = context.WithCancel(context.Background())

	server.Addoroutine(1)
	return server
}

func (server *Server) Addoroutine(num int) {
	for i := 0; i < num; i++ {
		atomic.AddInt32(&server.coNumber, 1)
		server.waitGroup.Add(1)
		cNum := server.coNumber
		go server.start(cNum)
	}
}

func (server *Server) IsStop() bool {
	return atomic.LoadInt32(&server.done) == 1
}

func (server *Server) callBack(msg *Msg) {
	objMethod := server.objMethod[msg.method]
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
			argsValue[0] = server.objValue
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
			argsValue[0] = server.objValue
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
			argsValue[0] = server.objValue
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
			argsValue[0] = server.objValue
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

func (server *Server) start(index int32) {
	defer server.waitGroup.Done()
	for {
		select {
		case <-server.ctx.Done():
			atomic.StoreInt32(&server.done, 1)
			if atomic.LoadInt32(&server.recvNumber) == atomic.LoadInt32(&server.sendNumber) {
				return
			}

		case cache := <-server.cache:
			msg := cache.(*Msg)
			atomic.AddInt32(&server.recvNumber, 1)
			server.callBack(msg)
			if atomic.LoadInt32(&server.done) == 1 &&
				atomic.LoadInt32(&server.recvNumber) == atomic.LoadInt32(&server.sendNumber) {
				return
			}
		}
	}
}

func (server *Server) Stop() {
	server.cancel()
	server.waitGroup.Wait()

}

func (server *Server) object() interface{} {
	return server.obj
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
	if objM := server.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
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
	if objM := server.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
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
	if objM := server.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
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
	if objM := server.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
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
