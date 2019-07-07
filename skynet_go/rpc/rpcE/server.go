package rpcE

import (
	"reflect"
	"skp-go/skynet_go/encodes"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"strings"
	"sync/atomic"
)

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
	service, err := NewService(obj)
	if err != nil {
		return nil
	}

	server := &Server{}
	server.service = service
	server.Server = rpc.NewServer(server)

	return server
}

func (server *Server) Object() interface{} {
	return server.service.obj
}

func (server *Server) ObjectName() string {
	return server.service.objName
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

func (server *Server) getFuncValue(objMethod *ObjMethod, msg *rpc.Msg) (reflect.Value, reflect.Value, error) {
	var args reflect.Value
	// Decode the argument value.
	argIsValue := false // if true, need to indirect before calling.
	if objMethod.argType.Kind() == reflect.Ptr {
		args = reflect.New(objMethod.argType.Elem())
	} else {
		args = reflect.New(objMethod.argType)
		argIsValue = true
	}

	err := encodes.DecodeBody(msg.Encode, msg.Args.(string), args.Interface())
	if err != nil {
		return reflect.Value{}, reflect.Value{}, err
	}

	//buf := bytes.NewBufferString()
	// argv guaranteed to be a pointer now.
	if argIsValue {
		args = args.Elem()
	}

	reply := reflect.New(objMethod.replyType.Elem())

	switch objMethod.replyType.Elem().Kind() {
	case reflect.Map:
		reply.Elem().Set(reflect.MakeMap(objMethod.replyType.Elem()))
	case reflect.Slice:
		reply.Elem().Set(reflect.MakeSlice(objMethod.replyType.Elem(), 0, 0))
	}
	return args, reply, nil
}

func (server *Server) RPC_DoMsg(msg *rpc.Msg) {
	service := server.service
	objMethod := service.objMethod[msg.Method]

	if msg.Typ == rpc.TypSend {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "objName = %s, Method = %s, %+v", server.service.objName, msg.Method, err))
				}
			}()

			args, replyv, err := server.getFuncValue(objMethod, msg)
			if err != nil {
				log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
				return
			}

			// valuePool := server.valuePools[2]
			// argsValue := valuePool.Get().([]reflect.Value)
			// argsValue[0] = server.service.objValue
			// argsValue[1] = args
			// argsValue[2] = replyv
			// retValues := objMethod.method.Func.Call(argsValue)

			objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})

		}()
		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypSendReq {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "objName = %s, Method = %s, %+v", server.service.objName, msg.Method, err))
				}
			}()

			args, replyv, err := server.getFuncValue(objMethod, msg)
			if err == nil {
				// valuePool := server.valuePools[2]
				// argsValue := valuePool.Get().([]reflect.Value)
				// argsValue[0] = server.service.objValue
				// argsValue[1] = args
				// argsValue[2] = replyv
				// retValues := objMethod.method.Func.Call(argsValue)

				retValues := objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})
				errI := retValues[0].Interface()
				if errI != nil {
					msg.Err = errI
					log.ErrorCode(errorCode.NewErrCode(0, msg.Err.(error).Error()))
					msg.CB2("", msg.Err.(error))
				} else {
					msg.Reply, msg.Err = encodes.EncodeBody(msg.Encode, replyv.Interface())
					if msg.Err != nil {
						log.ErrorCode(errorCode.NewErrCode(0, msg.Err.(error).Error()))
						msg.CB2("", msg.Err.(error))
					} else {
						msg.CB2(msg.Reply.(string), nil)
					}
				}
			} else {
				log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
				msg.CB2("", err)
			}
		}()

		server.MsgPool.Put(msg)
	} else if msg.Typ == rpc.TypCall {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorCode(errorCode.NewErrCode(0, "objName = %s, Method = %s, %+v", server.service.objName, msg.Method, err))
				}
			}()

			args, replyv, err := server.getFuncValue(objMethod, msg)
			if err == nil {
				// valuePool := server.valuePools[2]
				// argsValue := valuePool.Get().([]reflect.Value)
				// argsValue[0] = server.service.objValue
				// argsValue[1] = args
				// argsValue[2] = replyv
				// retValues := objMethod.method.Func.Call(argsValue)

				retValues := objMethod.method.Func.Call([]reflect.Value{server.service.objValue, args, replyv})
				errI := retValues[0].Interface()
				if errI != nil {
					msg.Err = errI.(error)
				} else {
					msg.Reply, msg.Err = encodes.EncodeBody(msg.Encode, replyv.Interface())
				}
			} else {
				log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
				msg.Err = err
			}

		}()
		msg.Pending <- msg
	}
}

func (server *Server) Send(encode int32, method string, args string) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSend
	msg.Method = method
	msg.Args = args
	msg.Encode = encode

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	return nil
}

func (server *Server) SendReq(encode int32, method string, args string, callBack rpc.CallBack2) error {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	msg.Typ = rpc.TypSendReq
	msg.Method = method
	msg.Args = args
	msg.CB2 = callBack
	msg.Encode = encode

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	return nil
}

func (server *Server) Call(encode int32, method string, args string) (string, error) {
	service := server.service
	if objM := service.objMethod[method]; objM == nil {
		return "", log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", service.objName, method))
	}

	msg := server.MsgPool.Get().(*rpc.Msg)
	msg.Init()
	defer server.MsgPool.Put(msg)
	msg.Typ = rpc.TypCall
	msg.Method = method
	msg.Args = args
	msg.Encode = encode

	atomic.AddInt32(&server.SendNumber, 1)
	server.Cache <- msg
	<-msg.Pending

	if msg.Err != nil {
		return "", msg.Err.(error)
	}

	return msg.Reply.(string), nil
}
