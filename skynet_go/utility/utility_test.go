package utility

import (
	log "skp-go/skynet_go/logger"
	"testing"
)

func Test_Int(t *testing.T) {
	var err error
	log.SetLevel(log.Lerr)
	in := 10
	out := 0
	b, err := IntToBytes(in)
	if err != nil {
		t.Error()
	}
	out, err = BytesToInt(b)
	if err != nil {
		t.Error()
	}
	if in != out {
		t.Error()
	}
}
