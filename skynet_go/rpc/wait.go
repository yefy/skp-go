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

type WaitCallBack func()
type WaitCallBack2 func() bool

func Timer(t time.Duration, callBack WaitCallBack2) {
	timer := time.NewTimer(t)
	for {
		select {
		case <-timer.C:
			if callBack() {
				return
			}
		}

		timer.Reset(t)
	}
}
