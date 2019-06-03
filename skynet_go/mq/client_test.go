package mq

import (
	log "skp-go/skynet_go/logger"
	"testing"
	"time"
)

func Test_Client1(t *testing.T) {
	log.SetLevel(log.Lerr)
	mqClient := NewClient("Test", ":5678")
	mqClient.Subscribe("Test", "*")
	mqClient.Start()
	time.Sleep(3 * time.Second)
}
