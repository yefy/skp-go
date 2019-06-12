package rpc

import (
	"testing"
)

func Test_Timer(t *testing.T) {
	n := 0
	w := NewWait()
	w.Timer(func() bool {
		n++
		if n < 10 {
			return false
		} else {
			return true
		}
	})
}
