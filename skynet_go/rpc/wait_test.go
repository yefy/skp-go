package rpc

import (
	log "skp-go/skynet_go/logger"
	"testing"
	"time"
)

func Test_Timer(t *testing.T) {
	n := 0
	w := NewWait()
	w.Timer(time.Millisecond*100, func() bool {
		n++
		if n < 10 {
			log.Fatal("false")
			return false
		} else {
			log.Fatal("true")
			return true
		}
	})
}
