package utility

import (
	"container/list"
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

func (self *SyncList) Push(data interface{}) {
	self.Lock()
	defer self.Unlock()
	self.list.PushBack(data)
}

func (self *SyncList) Pop() interface{} {
	self.Lock()
	defer self.Unlock()
	data := self.list.Front()
	if data == nil {
		return nil
	}
	self.list.Remove(data)
	return data.Value
}
