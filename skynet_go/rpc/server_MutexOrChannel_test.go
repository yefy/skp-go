package rpc

import (
	"fmt"
	_ "skp-go/skynet_go/errorCode"
	_ "skp-go/skynet_go/logger"
	"sync"
	"testing"
	"time"
)

type Op struct {
	key int
	val int
}

var lock sync.Mutex

var m1 map[int]int
var m2 map[int]int
var max int = 500000
var doneMutex chan interface{}
var doneChan chan interface{}

func update_map_by_mutex(i int) {
	lock.Lock()
	m1[i] = i
	if len(m1) == max {
		fmt.Printf("%s mutex finish\n", time.Now())
		data := 1
		doneMutex <- data
	}
	lock.Unlock()
}

var ch chan Op

func update_map_by_chan(i int) {
	ch <- Op{key: i, val: i}
}

func wait_for_chan(m map[int]int) {
	for {
		select {
		case op := <-ch:
			m2[op.key] = op.val
			if len(m2) == max {
				fmt.Printf("%s chan finish\n", time.Now())
				data := 1
				doneChan <- data
				return
			}
		}
	}
}

func Test_MutexOrChannel(t *testing.T) {
	doneMutex = make(chan interface{}, 1)
	doneChan = make(chan interface{}, 1)

	m1 = make(map[int]int, max)
	m2 = make(map[int]int, max)
	ch = make(chan Op)
	go wait_for_chan(m2)
	for i := 0; i < max; i++ {
		go update_map_by_chan(i)
		go update_map_by_mutex(i)
	}
	<-doneMutex
	<-doneChan
	if false {
		t.Error()
	}
	// === RUN   Test_Main
	// 2019-05-19 21:25:05.2059403 +0800 CST m=+10.018200201 chan finish
	// 2019-05-19 21:25:06.7557944 +0800 CST m=+11.568054301 mutex finish
	// --- PASS: Test_Main (11.55s)
	// PASS
	// ok  	skp-go/skynet_go/rpc	12.909s
}
