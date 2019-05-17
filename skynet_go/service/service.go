package service

import (
	"container/list"
	"context"
	"fmt"
	"reflect"
	"sync"
)

type SyncList struct {
	sync.Mutex
	list *list.List
}

func NewSyncList() *SyncList {
	syncList := &SyncList{}
	syncList.list = list.New()
	return syncList
}

func (self *SyncList) push(data interface{}) {
	self.Lock()
	defer self.Unlock()
	self.list.PushBack(data)
}

func (self *SyncList) pop() interface{} {
	self.Lock()
	defer self.Unlock()
	data := self.list.Front()
	if data == nil {
		return nil
	}
	self.list.Remove(data)
	return data
}

var isTest bool = true

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

type Service struct {
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
}

func parseMethod(service *Service, obj interface{}) {
	service.objType = reflect.TypeOf(obj)
	service.objKind = service.objType.Kind()
	if service.objKind != reflect.Ptr {
		panic("service.objKind != reflect.Ptr")
	}
	service.objValue = reflect.ValueOf(obj)
	service.objName = reflect.Indirect(service.objValue).Type().Name()

	for i := 0; i < service.objType.NumMethod(); i++ {
		method := service.objType.Method(i)
		objM := &ObjMethod{}
		objM.method = method
		objM.typ = method.Type
		objM.name = method.Name // method.Func
		objM.value = service.objValue.MethodByName(method.Name)
		objM.argvIn = objM.value.Type().NumIn()
		objM.argvOut = objM.value.Type().NumOut()
		objM.argType = objM.typ.In(1)
		objM.replyType = objM.typ.In(2)
		service.objMethod[objM.name] = objM
		fmt.Printf("objM = %+v \n", objM)

		if objM.argvIn != 2 {
			panic("objM.argvIn != 2")
		}

		if objM.argvOut != 1 {
			panic("objM.argvOut != 1")
		}

		if objM.replyType.Kind() != reflect.Ptr {
			panic("objM.replyType.Kind() != reflect.Ptr")
		}

		objM.argName = objM.argType.Name()
		objM.replyName = objM.replyType.Elem().Name()
	}
}

func NewService(coNumber int, cacheNumber int, obj interface{}) *Service {
	fmt.Printf("NewService coNumber = %+v, cacheNumber = %+v \n", coNumber, cacheNumber)
	service := &Service{}
	service.objMethod = make(map[string]*ObjMethod)
	service.cacheNumber = cacheNumber
	service.cache = make(chan interface{}, cacheNumber)
	service.coNumber = coNumber

	parseMethod(service, obj)
	fmt.Printf("service = %+v \n", service)

	service.msgPool = &sync.Pool{New: func() interface{} {
		msg := &Msg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}

	service.waitGroup = &sync.WaitGroup{}
	service.ctx, service.cancel = context.WithCancel(context.Background())

	for i := 0; i < coNumber; i++ {
		service.waitGroup.Add(1)
		go service.start(i)
	}
	return service
}

func (self *Service) start(index int) {
	fmt.Println("Service start index = ", index)
	defer self.waitGroup.Done()
	for {
		select {
		case <-self.ctx.Done():
			fmt.Printf("Service stop index = %+v \n", index)
			return
		case msgInterface := <-self.cache:
			msg := msgInterface.(*Msg)
			fmt.Printf("msg = %+v \n", msg)
			objMethod := self.objMethod[msg.method]
			retErrs := objMethod.value.Call([]reflect.Value{reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})
			//objMethod.method.Func.Call([]reflect.Value{self.objValue, reflect.ValueOf(msg.args), reflect.ValueOf(msg.reply)})

			if msg.typ == "call" {
				msg.err = retErrs[0].Interface()
				msg.pending <- msg
			} else {
				self.msgPool.Put(msg)
			}
		}
	}
}

func (self *Service) Stop() {
	self.cancel()
	self.waitGroup.Wait()
	fmt.Printf("Service stop \n")
}

func (self *Service) checkArgv(objMethod *ObjMethod, args interface{}, reply interface{}) {
	if isTest == false {
		return
	}
	replyKind := reflect.TypeOf(reply).Kind()
	if replyKind != reflect.Ptr {
		panic(fmt.Sprintf("replyKind:%+v != reflect.Ptr \n", replyKind))
	}

	argName := reflect.TypeOf(args).Name()
	if argName != objMethod.argName {
		panic(fmt.Sprintf("argName:%+v != objMethod.argName:%+v \n", argName, objMethod.argName))
	}

	replyName := reflect.TypeOf(reply).Elem().Name()
	if replyName != objMethod.replyName {
		panic(fmt.Sprintf("argName:%+v != objMethod.argName:%+v \n", argName, objMethod.argName))
	}
}

func (self *Service) Send(funcName string, argv ...interface{}) {
	/*
		fmt.Printf("Service Send funcName = %+v \n", funcName)
		objM := self.objMethod[funcName]
		fmt.Printf("objM = %+v \n", objM)
		if objM == nil {
			panic(fmt.Sprintf("not funcName = %+v \n", funcName))
		}
		if len(argv) != objM.argvIn {
			panic(fmt.Sprintf("len(argv):%+v != objM.argvIn:%+v \n", len(argv), objM.argvIn))
		}

		msg := self.msgPool.Get().(*Msg)
		msg.funcName = funcName
		msg.argvs = argv
		msg.isRet = false

		self.cache <- msg
	*/
}

func (self *Service) Call(method string, args interface{}, reply interface{}) error {
	fmt.Printf("Service Call method =%+v \n", method)
	objM := self.objMethod[method]
	fmt.Printf("objM = %+v", objM)
	if objM == nil {
		panic(fmt.Sprintf("not method = %+v \n", method))
	}
	self.checkArgv(objM, args, reply)

	msg := self.msgPool.Get().(*Msg)
	msg.typ = "call"
	msg.method = method
	msg.args = args
	msg.reply = reply

	self.cache <- msg
	<-msg.pending
	err := msg.err
	self.msgPool.Put(msg)
	if err == nil {
		return nil
	}
	return err.(error)
}
