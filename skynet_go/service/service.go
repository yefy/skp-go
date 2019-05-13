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

type Msg struct {
	funcName string
	argvs    interface{}
	pending  chan interface{}
	isRet    bool
}

type Service struct {
	objName     string
	objMethod   map[string]*ObjMethod
	cacheNumber int
	cache       chan interface{}
	coNumber    int
	waitGroup   *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	msgPool     *sync.Pool
	funcMethod  map[uint64]*ObjMethod
}

func NewService(coNumber int, cacheNumber int, objFunc interface{}, argv ...interface{}) *Service {
	fmt.Printf("NewService coNumber = %+v, cacheNumber = %+v, len(argv) = %+v \n", coNumber, cacheNumber, len(argv))
	CheckKind(objFunc, reflect.Func)
	argvValue := GetArgvValue(argv...)
	funcM := GetMethod(objFunc)[0]
	retValue := funcM.value.Call(argvValue)
	if len(retValue) != 1 {
		panic("len(objectValues) != 1")
	}

	obj := retValue[0].Interface()
	CheckKind(obj, reflect.Ptr)

	service := &Service{}
	service.objName = GetObjName(obj)
	service.objMethod = make(map[string]*ObjMethod)
	service.cacheNumber = cacheNumber
	service.cache = make(chan interface{}, cacheNumber)
	service.coNumber = coNumber
	service.funcMethod = make(map[uint64]*ObjMethod)

	methods := GetMethod(obj)
	for _, value := range methods {
		service.objMethod[value.name] = value
	}

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
			argvs := msg.argvs.([]interface{})
			argvValue := GetArgvsValue(argvs)
			retValue := self.objMethod[msg.funcName].value.Call(argvValue)

			if msg.isRet {
				msg.pending <- retValue
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

func (self *Service) Send(funcName string, argv ...interface{}) {
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
}

func (self *Service) Call(retFunc interface{}, funcName string, argv ...interface{}) {
	fmt.Printf("Service Call retFunc = %+v, funcName =%+v \n", retFunc, funcName)
	CheckKind(retFunc, reflect.Func)
	funcAddr := uint64(reflect.ValueOf(retFunc).Pointer())
	funcM := self.funcMethod[funcAddr]
	if funcM == nil {
		funcM = GetMethod(retFunc)[0]
		self.funcMethod[funcAddr] = funcM
	}
	fmt.Printf("funcM = %+v \n", funcM)
	objM := self.objMethod[funcName]
	fmt.Printf("objM = %+v", objM)
	if objM == nil {
		panic(fmt.Sprintf("not funcName = %+v \n", funcName))
	}
	if len(argv) != objM.argvIn {
		panic(fmt.Sprintf("len(argv):%+v != objM.argvIn:%+v\n", len(argv), objM.argvIn))
	}

	if funcM.argvIn != objM.argvOut {
		panic(fmt.Sprintf("funcM.argvIn:%+v != objM.argvOut:%+v \n", funcM.argvIn, objM.argvOut))
	}

	msg := self.msgPool.Get().(*Msg)
	msg.funcName = funcName
	msg.argvs = argv
	msg.isRet = true

	self.cache <- msg

	retValue := (<-msg.pending).([]reflect.Value)
	if funcM.argvIn != len(retValue) {
		panic(fmt.Sprintf("funcM.argvIn:%+v != len(retValue):%+v \n", funcM.argvIn, len(retValue)))
	}
	self.msgPool.Put(msg)

	argvValue := GetInterfaceValue(retValue)
	funcM.value.Call(argvValue)
}
