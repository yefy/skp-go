package rpc

import (
	"sync"
	"time"
)

type Wait struct {
	waitGroup *sync.WaitGroup
}

func NewWait() *Wait {
	w := &Wait{}
	w.waitGroup = &sync.WaitGroup{}
	return w
}

type WaitCallBack func()
type WaitCallBack2 func() bool

func (w *Wait) Run(callBack WaitCallBack) {
	w.waitGroup.Add(1)
	go func() {
		callBack()
		w.waitGroup.Done()
	}()
}

func (w *Wait) Wait() {
	w.waitGroup.Wait()
}

func (w *Wait) Timer(callBack WaitCallBack2) {
	t := time.Duration(100) * time.Millisecond
	timer := time.NewTimer(t)
	isExit := false
	for {
		select {
		case <-timer.C:
			isExit = callBack()
		}

		if isExit {
			break
		} else {
			timer.Reset(t)
		}
	}
}
