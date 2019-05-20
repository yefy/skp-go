package rpc

import (
	"context"
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	_ "skp-go/skynet_go/utility"
	"sync"
	"sync/atomic"
)

var isTest bool = false

type Msg struct {
	typ     string //call send
	method  string
	args    interface{}
	reply   interface{}
	pending chan interface{}
	err     interface{}
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
	coNumber    int
	waitGroup   *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	msgPool     *sync.Pool
	//msgPool    *utility.SyncPool
	sendNumber int32
	recvNumber int32
	done       int32
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
		objM.name = method.Name // method.Func
		objM.value = server.objValue.MethodByName(method.Name)
		objM.argvIn = objM.value.Type().NumIn()
		objM.argvOut = objM.value.Type().NumOut()
		objM.argType = objM.typ.In(1)
		objM.replyType = objM.typ.In(2)
		server.objMethod[objM.name] = objM

		if objM.argvIn != 2 {
			return log.Panic(errorCode.NewErrCode(0, "objM.argvIn != 2, name = %+v, method = %+v", server.objName, objM.name))
		}

		if objM.argvOut != 1 {
			return log.Panic(errorCode.NewErrCode(0, "objM.argvOut != 1, name = %+v, method = %+v", server.objName, objM.name))
		}

		if objM.replyType.Kind() != reflect.Ptr {
			return log.Panic(errorCode.NewErrCode(0, "objM.replyType.Kind() != reflect.Ptr, name = %+v, method = %+v", server.objName, objM.name))
		}

		objM.argName = objM.argType.Name()
		objM.replyName = objM.replyType.Elem().Name()

		//log.All("objM = %+v", objM)
	}
	return nil
}

func NewServer(coNumber int, cacheNumber int, obj interface{}) *Server {
	//log.All("NewServer coNumber = %+v, cacheNumber = %+v", coNumber, cacheNumber)
	server := &Server{}
	server.obj = obj
	server.objMethod = make(map[string]*ObjMethod)
	server.cacheNumber = cacheNumber
	server.cache = make(chan interface{}, cacheNumber)
	server.coNumber = coNumber
	server.sendNumber = 0
	server.recvNumber = 0
	server.done = 0

	err := parseMethod(server, obj)
	if err != nil {
		return nil
	}

	//log.All("server = %+v", server)

	server.msgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}

	// server.msgPool = utility.NewSyncPool(func() interface{} {
	// 	msg := &Msg{}
	// 	msg.pending = make(chan interface{}, 1)
	// 	return msg
	// })

	server.waitGroup = &sync.WaitGroup{}
	server.ctx, server.cancel = context.WithCancel(context.Background())

	for i := 0; i < coNumber; i++ {
		server.waitGroup.Add(1)
		go server.start(i)
	}
	return server
}

func (server *Server) start(index int) {
	//log.All("Server %+v start index = %+v", server.objName, index)
	defer server.waitGroup.Done()
	for {
		select {
		case <-server.ctx.Done():
			//log.Fatal("Server %+v stop index = %+v", server.objName, index)
			atomic.StoreInt32(&server.done, 1)
			if atomic.LoadInt32(&server.recvNumber) == atomic.LoadInt32(&server.sendNumber) {
				return
			}

		case cache := <-server.cache:
			msg := cache.(*Msg)
			//log.All("%+v msg = %+v", server.objName, msg)

			objMethod := server.objMethod[msg.method]
			//retErrs := objMethod.value.Call([]reflect.Value{reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})
			retErrs := objMethod.method.Func.Call([]reflect.Value{server.objValue, reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})

			if msg.typ == "call" {
				msg.err = retErrs[0].Interface()
				msg.pending <- msg
			} else {
				server.msgPool.Put(msg)
			}
			atomic.AddInt32(&server.recvNumber, 1)
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
	//log.Fatal("Server %+v stop", server.objName)
}

func (server *Server) checkArgv(objMethod *ObjMethod, args interface{}, reply interface{}) error {
	if isTest == false {
		return nil
	}
	replyKind := reflect.TypeOf(reply).Kind()
	if replyKind != reflect.Ptr {
		return log.Panic(errorCode.NewErrCode(0, "%+v : replyKind:%+v != reflect.Ptr", server.objName, replyKind))
	}

	argName := reflect.TypeOf(args).Name()
	if argName != objMethod.argName {
		return log.Panic(errorCode.NewErrCode(0, "%+v : argName:%+v != objMethod.argName:%+v", server.objName, argName, objMethod.argName))
	}

	replyName := reflect.TypeOf(reply).Elem().Name()
	if replyName != objMethod.replyName {
		return log.Panic(errorCode.NewErrCode(0, "%+v : argName:%+v != objMethod.argName:%+v", server.objName, argName, objMethod.argName))
	}
	return nil
}

func (server *Server) Send(method string, args interface{}, reply interface{}) error {
	//log.All("Server %+v Send method =%+v", server.objName, method)
	objM := server.objMethod[method]
	if objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
	}
	//log.All("objM = %+v", objM)
	err := server.checkArgv(objM, args, reply)
	if err != nil {
		return err
	}

	msg := server.msgPool.Get().(*Msg)
	msg.typ = "send"
	msg.method = method
	msg.args = args
	msg.reply = reply

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	return nil
}

func (server *Server) Call(method string, args interface{}, reply interface{}) error {
	//log.All("Server %+v Call method =%+v", server.objName, method)
	objM := server.objMethod[method]
	if objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
	}
	//log.All("objM = %+v", objM)
	err := server.checkArgv(objM, args, reply)
	if err != nil {
		return err
	}

	msg := server.msgPool.Get().(*Msg)
	msg.typ = "call"
	msg.method = method
	msg.args = args
	msg.reply = reply

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	<-msg.pending
	msgErr := msg.err
	server.msgPool.Put(msg)
	if msgErr == nil {
		return nil
	}
	return msgErr.(error)
}
