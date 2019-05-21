package rpc_gob

import (
	"bytes"
	"context"
	"encoding/gob"
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	_ "skp-go/skynet_go/utility"
	"sync"
	"sync/atomic"
)

var isTest bool = false

func SetTest(b bool) {
	isTest = b
}

type Msg struct {
	typ       string //call send
	method    string
	args      interface{}
	reply     interface{}
	pending   chan interface{}
	err       interface{}
	replyFunc interface{}
	argsStr   string
	replyStr  string
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
		if objM.argvIn > 0 {
			objM.argType = objM.typ.In(1)
			objM.argName = objM.argType.Name()
		}
		if objM.argvIn > 1 {
			objM.replyType = objM.typ.In(2)
			if objM.replyType.Kind() != reflect.Ptr {
				return log.Panic(errorCode.NewErrCode(0, "objM.replyType.Kind() != reflect.Ptr, name = %+v, method = %+v", server.objName, objM.name))
			}

			objM.replyName = objM.replyType.Elem().Name()
		}

		if objM.argvOut > 0 {
			if objM.typ.Out(0).Name() != "error" {
				return log.Panic(errorCode.NewErrCode(0, "objM.typ.Out(0).Name() != error, name = %+v, method = %+v", server.objName, objM.name))
			}
		}

		if objM.argvIn > 2 {
			return log.Panic(errorCode.NewErrCode(0, "objM.argvIn > 2, name = %+v, method = %+v", server.objName, objM.name))
		}

		if objM.argvOut > 1 {
			return log.Panic(errorCode.NewErrCode(0, "objM.argvOut > 1, name = %+v, method = %+v", server.objName, objM.name))
		}

		server.objMethod[objM.name] = objM
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

func (server *Server) callBack(msg *Msg) {
	objMethod := server.objMethod[msg.method]
	if msg.typ == "call" {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Err("%+v", err)
					msg.err = errorCode.NewErrCode(0, "%+v", err)
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
			buf.WriteString(msg.argsStr)
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

			retValues := objMethod.method.Func.Call([]reflect.Value{server.objValue, reflect.ValueOf(args.Interface()), reflect.ValueOf(replyv.Interface())})
			msg.err = retValues[0].Interface()

			var bufOut bytes.Buffer
			enc := gob.NewEncoder(&bufOut)
			enc.Encode(replyv.Interface())
			msg.replyStr = bufOut.String()

		}()
		msg.pending <- msg
	} else if msg.typ == "send" {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Err("%+v", err)
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
			buf.WriteString(msg.argsStr)
			dec := gob.NewDecoder(&buf)
			dec.Decode(args.Interface())
			//buf := bytes.NewBufferString()
			// argv guaranteed to be a pointer now.
			if argIsValue {
				args = args.Elem()
			}
			objMethod.method.Func.Call([]reflect.Value{server.objValue, reflect.ValueOf(args.Interface())})

		}()
		server.msgPool.Put(msg)
	} else if msg.typ == "asynCall" {
		var retValues []reflect.Value
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Err("%+v", err)
					retValues = []reflect.Value{reflect.ValueOf(errorCode.NewErrCode(0, "%+v", err))}
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
			var bufIn bytes.Buffer
			bufIn.WriteString(msg.argsStr)
			dec := gob.NewDecoder(&bufIn)
			dec.Decode(args.Interface())
			//bufIn := bytes.NewBufferString()
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

			retValues = objMethod.method.Func.Call([]reflect.Value{server.objValue, reflect.ValueOf(args.Interface()), reflect.ValueOf(replyv.Interface())})
			var bufOut bytes.Buffer
			enc := gob.NewEncoder(&bufOut)
			enc.Encode(replyv.Interface())

			var bufOut2 bytes.Buffer
			bufOut2.WriteString(bufOut.String())
			dec2 := gob.NewDecoder(&bufOut2)
			dec2.Decode(msg.reply)

		}()
		funcValue := reflect.ValueOf(msg.replyFunc)
		funcValue.Call(retValues)
		server.msgPool.Put(msg)
	}
	//retErrs := objMethod.value.Call([]reflect.Value{reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})
	//retErrs := objMethod.method.Func.Call([]reflect.Value{server.objValue, reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})
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
			server.callBack(msg)
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

	argName := reflect.TypeOf(args).Name()
	if argName != objMethod.argName {
		return log.Panic(errorCode.NewErrCode(0, "%+v : argName:%+v != objMethod.argName:%+v", server.objName, argName, objMethod.argName))
	}

	if reply != nil {
		replyKind := reflect.TypeOf(reply).Kind()
		if replyKind != reflect.Ptr {
			return log.Panic(errorCode.NewErrCode(0, "%+v : replyKind:%+v != reflect.Ptr", server.objName, replyKind))
		}

		replyName := reflect.TypeOf(reply).Elem().Name()
		if replyName != objMethod.replyName {
			return log.Panic(errorCode.NewErrCode(0, "%+v : argName:%+v != objMethod.argName:%+v", server.objName, argName, objMethod.argName))
		}
	}
	return nil
}

func (server *Server) Send(method string, args interface{}) error {
	//log.All("Server %+v Send method =%+v", server.objName, method)
	objM := server.objMethod[method]
	if objM == nil {
		return log.Panic(errorCode.NewErrCode(0, "%+v not method = %+v", server.objName, method))
	}
	//log.All("objM = %+v", objM)
	err := server.checkArgv(objM, args, nil)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(args)

	msg := server.msgPool.Get().(*Msg)
	msg.typ = "send"
	msg.method = method
	msg.args = args
	msg.argsStr = buf.String()

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	return nil
}

func (server *Server) asynCall(method string, args interface{}, reply interface{}, replyFunc interface{}) error {
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

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(args)

	msg := server.msgPool.Get().(*Msg)
	msg.typ = "asynCall"
	msg.method = method
	msg.args = args
	msg.argsStr = buf.String()
	msg.reply = reply
	msg.replyFunc = replyFunc

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

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(args)

	msg := server.msgPool.Get().(*Msg)
	msg.typ = "call"
	msg.method = method
	msg.args = args
	msg.argsStr = buf.String()
	msg.reply = reply

	atomic.AddInt32(&server.sendNumber, 1)
	server.cache <- msg
	<-msg.pending

	var bufOut bytes.Buffer
	bufOut.WriteString(msg.replyStr)
	dec := gob.NewDecoder(&bufOut)
	dec.Decode(reply)

	msgErr := msg.err
	server.msgPool.Put(msg)
	if msgErr == nil {
		return nil
	}
	return msgErr.(error)
}
