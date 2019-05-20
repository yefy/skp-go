package utility

import (
	"container/list"
	"sync"
)

type SyncPool struct {
	sync.Mutex
	list   *list.List
	create func() interface{}
}

func NewSyncPool(create func() interface{}) *SyncPool {
	syncList := &SyncPool{}
	syncList.list = list.New()
	syncList.create = create
	return syncList
}

func (self *SyncPool) Put(data interface{}) {
	self.Lock()
	defer self.Unlock()
	self.list.PushBack(data)
}

func (self *SyncPool) Get() interface{} {
	self.Lock()
	defer self.Unlock()
	data := self.list.Front()
	if data == nil {
		return self.create()
	}
	self.list.Remove(data)
	return data.Value
}
