package skpSocket

import (
	"sync"
)

type SkpObject struct {
	ObjectCountMutex sync.Mutex
	ObjectCount      int32
}

func SkpNewObject() (this *SkpObject) {
	object := new(SkpObject)
	object.ObjectCount = 0
	return object
}

func (this *SkpObject) SkpObjectCountLock() {
	this.ObjectCountMutex.Lock()
}

func (this *SkpObject) SkpObjectCountUnlock() {
	this.ObjectCountMutex.Unlock()
}

func (this *SkpObject) SkpObjectCount() int32 {
	return this.ObjectCount
}

func (this *SkpObject) SkpObjectCountIncr() int32 {
	this.ObjectCount++
	return this.ObjectCount
}

func (this *SkpObject) SkpObjectCountDecr() int32 {
	this.ObjectCount--
	return this.ObjectCount
}

func (this *SkpObject) SkpObjectCountIncrAtomic() int32 {
	this.SkpObjectCountLock()
	defer this.SkpObjectCountUnlock()

	this.ObjectCount++
	return this.ObjectCount
}

func (this *SkpObject) SkpObjectCountDecrAtomic() int32 {
	this.SkpObjectCountLock()
	defer this.SkpObjectCountUnlock()

	this.ObjectCount--
	return this.ObjectCount
}
