package rpcGob

import (
	"bytes"
	"context"
	"encoding/gob"
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"strings"
	"sync"
	"sync/atomic"
)

var isTest bool = true

func SetTest(b bool) {
	isTest = b
}

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

type CallBack func(string, error)

type Msg struct {
	typ      string
	method   string
	args     string
	reply    string
	pending  chan *Msg
	err      error
	callBack CallBack
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

		if strings.Contains(objM.name, "RPC_") {
			continue
		}

		if objM.argvIn > 0 {
			objM.argType = objM.typ.In(1)
			objM.argName = objM.argType.Name()
		}
		if objM.argvIn > 1 {
			objM.replyType = objM.typ.In(2)
			if objM.replyType.Kind() != reflect.Ptr {
				return nil, log.Panic(errorCode.NewErrCode(0, "objM.replyType.Kind() != reflect.Ptr, name = %+v, method = %+v", s.objName, objM.name))
			}

			objM.replyName = objM.replyType.Elem().Name()
		}

		if objM.argvOut > 0 {
			if objM.typ.Out(0).Name() != "error" {
				return nil, log.Panic(errorCode.NewErrCode(0, "objM.typ.Out(0).Name() != error, name = %+v, method = %+v", s.objName, objM.name))
			}
		}

		if objM.argvIn > 2 {
			return nil, log.Panic(errorCode.NewErrCode(0, "objM.argvIn > 2, name = %+v, method = %+v", s.objName, objM.name))
		}

		if objM.argvOut > 1 {
			return nil, log.Panic(errorCode.NewErrCode(0, "objM.argvOut > 1, name = %+v, method = %+v", s.objName, objM.name))
		}

		s.objMethod[objM.name] = objM
	}

	return s, nil
}

type ServerInterface interface {
	RPC_SetServer(*Server)
	RPC_GetServer() *Server
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
	mutex           sync.Mutex
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
		msg.pending = make(chan *Msg, 1)
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

	obj.RPC_SetServer(server)

	server.Start(true)

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
		server.service.obj.RPC_Close()
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
	objMethod := service.objMethod[msg.method]

	if msg.typ == typSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			var args reflect.Value
			// Decode the argument value.
			argIsValue := false // if true, need to indirect before calling.
			if objMethod.argType.Kind() == reflect.Ptr {
				args = reflect.New(objMethod.argType.Elem())
			} else {
				args = reflect.New(objMethod.argType)
				argIsValue = true
			}

			var buf bytes.Buffer
			buf.WriteString(msg.args)
			dec := gob.NewDecoder(&buf)
			dec.Decode(args.Interface())
			//buf := bytes.NewBufferString()
			// argv guaranteed to be a pointer now.
			if argIsValue {
				args = args.Elem()
			}

			replyv := reflect.New(objMethod.replyType.Elem())

			switch objMethod.replyType.Elem().Kind() {
			case reflect.Map:
				replyv.Elem().Set(reflect.MakeMap(objMethod.replyType.Elem()))
			case reflect.Slice:
				replyv.Elem().Set(reflect.MakeSlice(objMethod.replyType.Elem(), 0, 0))
			}

			// valuePool := server.valuePools[2]
			// argsValue := valuePool.Get().([]reflect.Value)
			// argsValue[0] = server.service.objValue
			// argsValue[1] = args
			// argsValue[2] = replyv
			// retValues := objMethod.method.Func.Call(argsValue)

			objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})

		}()
		server.msgPool.Put(msg)
	} else if msg.typ == typSendReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			var args reflect.Value
			// Decode the argument value.
			argIsValue := false // if true, need to indirect before calling.
			if objMethod.argType.Kind() == reflect.Ptr {
				args = reflect.New(objMethod.argType.Elem())
			} else {
				args = reflect.New(objMethod.argType)
				argIsValue = true
			}

			var buf bytes.Buffer
			buf.WriteString(msg.args)
			dec := gob.NewDecoder(&buf)
			dec.Decode(args.Interface())
			//buf := bytes.NewBufferString()
			// argv guaranteed to be a pointer now.
			if argIsValue {
				args = args.Elem()
			}

			replyv := reflect.New(objMethod.replyType.Elem())

			switch objMethod.replyType.Elem().Kind() {
			case reflect.Map:
				replyv.Elem().Set(reflect.MakeMap(objMethod.replyType.Elem()))
			case reflect.Slice:
				replyv.Elem().Set(reflect.MakeSlice(objMethod.replyType.Elem(), 0, 0))
			}

			// valuePool := server.valuePools[2]
			// argsValue := valuePool.Get().([]reflect.Value)
			// argsValue[0] = server.service.objValue
			// argsValue[1] = args
			// argsValue[2] = replyv
			// retValues := objMethod.method.Func.Call(argsValue)

			retValues := objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})
			errI := retValues[0].Interface()
			if errI != nil {

				msg.err = errI.(error)
			} else {
				msg.err = nil
			}

			var bufOut bytes.Buffer
			enc := gob.NewEncoder(&bufOut)
			enc.Encode(replyv.Interface())
			msg.reply = bufOut.String()
			msg.callBack(msg.reply, msg.err)
		}()

		server.msgPool.Put(msg)
	} else if msg.typ == typCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.err = log.ErrorCode(errorCode.NewErrCode(0, "%+v", err))
				}
			}()

			var args reflect.Value
			// Decode the argument value.
			argIsValue := false // if true, need to indirect before calling.
			if objMethod.argType.Kind() == reflect.Ptr {
				args = reflect.New(objMethod.argType.Elem())
			} else {
				args = reflect.New(objMethod.argType)
				argIsValue = true
			}

			var buf bytes.Buffer
			buf.WriteString(msg.args)
			dec := gob.NewDecoder(&buf)
			dec.Decode(args.Interface())
			//buf := bytes.NewBufferString()
			// argv guaranteed to be a pointer now.
			if argIsValue {
				args = args.Elem()
			}

			replyv := reflect.New(objMethod.replyType.Elem())

			switch objMethod.replyType.Elem().Kind() {
			case reflect.Map:
				replyv.Elem().Set(reflect.MakeMap(objMethod.replyType.Elem()))
			case reflect.Slice:
				replyv.Elem().Set(reflect.MakeSlice(objMethod.replyType.Elem(), 0, 0))
			}

			// valuePool := server.valuePools[2]
			// argsValue := valuePool.Get().([]reflect.Value)
			// argsValue[0] = server.service.objValue
			// argsValue[1] = args
			// argsValue[2] = replyv
			// retValues := objMethod.method.Func.Call(argsValue)

			retValues := objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})
			errI := retValues[0].Interface()
			if errI != nil {

				msg.err = errI.(error)
			} else {
				msg.err = nil
			}

			var bufOut bytes.Buffer
			enc := gob.NewEncoder(&bufOut)
			enc.Encode(replyv.Interface())
			msg.reply = bufOut.String()
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

func (server *Server) Send(method string, args string) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
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

func (server *Server) SendReq(method string, args string, callBack CallBack) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	msg := server.msgPool.Get().(*Msg)
	msg.init()
	msg.typ = typSendReq
	msg.method = method
	msg.args = args
	msg.callBack = callBack

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	return nil
}

func (server *Server) Call(method string, args string) (string, error) {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return "", log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
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
		return "", msg.err
	}

	return msg.reply, nil
}
