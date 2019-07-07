package rpcU

import (
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync/atomic"
)

var isCheck bool = true

func SetCheck(b bool) {
	isCheck = b
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
	obj       ServerI
	objName   string
	objType   reflect.Type
	objValue  reflect.Value
	objKind   reflect.Kind
	objMethod map[string]*ObjMethod
}

func NewService(obj ServerI) (*Service, error) {
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

type ServerI interface {
	RPC_Describe() string
	RPC_SetServer(*Server)
	RPC_GetServer() *Server
	RPC_Stop()
}

type ServerB struct {
	server *Server
}

func (sb *ServerB) RPC_SetServer(server *Server) {
	if sb.server != nil {
		log.Panic(errorCode.NewErrCode(0, server.RPC_Describe()))
	}
	sb.server = server
}

func (sb *ServerB) RPC_GetServer() *Server {
	return sb.server
}

func (sb *ServerB) RPC_Stop() {

}

type Server struct {
	service *Service
	*rpc.Server
}

func NewServer(obj ServerI) *Server {
	log.Fatal("NewServer 0000000000")
	service, err := NewService(obj)
	if err != nil {
		return nil
	}
	log.Fatal("NewServer 11111111111")
	server := &Server{}
	server.service = service
	server.Server = rpc.NewServer(server)
	log.Fatal("NewServer 2222222222")
	return server
}
func (server *Server) Object() interface{} {
	return server.service.obj
}

func (server *Server) RPC_Describe() string {
	return server.service.obj.RPC_Describe()
}

func (server *Server) RPC_Start() {
	server.service.obj.RPC_SetServer(server)
}

func (server *Server) RPC_Stop() {
	server.service.obj.RPC_Stop()
}

func (server *Server) RPC_DoMsg(msg *rpc.Msg) {
	service := server.service
	objMethod := service.objMethod[msg.Method]
	args := msg.Args.([]interface{})

	if msg.Typ == rpc.TypSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "Typ = %d, objName = %s, Method = %s, %+v", msg.Typ, server.service.objName, msg.Method, err))
					log.Err("service.objValue = %+v", service.objValue)
					for i := 0; i < len(args); i++ {
						log.Err("i = %d, args = %+v", i, args[i])
					}
				}
			}()
			valuePool := server.ValuePools[len(args)]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args); i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			objMethod.method.Func.Call(argsValue)
		}()
		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypSendReq {
		var replyValues []reflect.Value
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "Typ = %d, objName = %s, Method = %s, %+v", msg.Typ, server.service.objName, msg.Method, err))
					log.Err("service.objValue = %+v", service.objValue)
					for i := 0; i < len(args); i++ {
						log.Err("i = %d, args = %+v", i, args[i])
					}
				}
			}()
			valuePool := server.ValuePools[len(args)-1]
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

		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "Typ = %d, objName = %s, Method = %s, %+v", msg.Typ, server.service.objName, msg.Method, err))
					log.Err("service.objValue = %+v", service.objValue)
					for i := 0; i < len(args); i++ {
						log.Err("i = %d, args = %+v", i, args[i])
					}
				}
			}()

			valuePool := server.ValuePools[len(args)]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args); i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			replyValues := objMethod.method.Func.Call(argsValue)
			msg.Reply = replyValues
		}()
		msg.Pending <- msg
	} else if msg.Typ == rpc.TypCallReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "Typ = %d, objName = %s, Method = %s, %+v", msg.Typ, server.service.objName, msg.Method, err))
					log.Err("service.objValue = %+v", service.objValue)
					for i := 0; i < len(args); i++ {
						log.Err("i = %d, args = %+v", i, args[i])
					}
				}
			}()

			valuePool := server.ValuePools[len(args)-1]
			argsValue := valuePool.Get().([]reflect.Value)
			argsValue[0] = service.objValue
			for i := 0; i < len(args)-1; i++ {
				argsValue[i+1] = reflect.ValueOf(args[i])
			}

			replyValues := objMethod.method.Func.Call(argsValue)
			msg.Reply = replyValues
		}()
		msg.Pending <- msg
	}
	//objMethod.value.Call([]reflect.Value{reflect.ValueOf(args)})
	//objMethod.method.Func.Call([]reflect.Value{objValue, reflect.ValueOf(args)})
}

func checkArgv(isFunc bool, args ...interface{}) error {
	if isCheck == false {
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

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSend
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
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

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSendReq
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
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

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	defer server.MsgPool.Put(msg)
	msg.Typ = rpc.TypCall
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	<-msg.Pending

	if msg.Err != nil {
		return msg.Err.(error)
	}

	replyValues := msg.Reply.([]reflect.Value)
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

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	defer server.MsgPool.Put(msg)
	msg.Typ = rpc.TypCallReq
	msg.Method = method
	msg.Args = args

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	<-msg.Pending

	if msg.Err == nil {
		replyFunc := args[len(args)-1]
		replyFuncName := reflect.ValueOf(replyFunc)
		func() {
			defer func() {
				if err := recover(); err != nil {
					msg.Err = log.Panic(errorCode.NewErrCode(0, "%+v", err))
				}
			}()
			replyFuncName.Call(msg.Reply.([]reflect.Value))
		}()
	}

	if msg.Err != nil {
		return msg.Err.(error)
	}

	return nil
}
