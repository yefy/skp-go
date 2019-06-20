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

func Run(callBack WaitCallBack) {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		callBack()
		waitGroup.Done()
	}()
	waitGroup.Wait()
}

func Timer(t time.Duration, callBack WaitCallBack2) {
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
