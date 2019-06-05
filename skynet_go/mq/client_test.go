package mq

import (
	log "skp-go/skynet_go/logger"
	"testing"
	"time"
)

func Test_Client1(t *testing.T) {
	log.SetLevel(log.Lerr)
	mqClient := NewClient("Test", ":5671")
	mqClient.Subscribe("Test", "*")
	mqClient.Start()
	time.Sleep(2 * time.Second)
	request := RegisteRequest{}
	request.Instance = "Instance"
	request.Harbor = 1
	request.Topic = "Topic"
	request.Tag = "Tag"
	reply := RegisterReply{}
	msg := Msg{Topic: "Test", Tag: "0"}
	if err := mqClient.Call(&msg, "Test.OnRegister", &request, &reply); err != nil {
		t.Error()
	}

	log.Fatal("creply.Harbor = %d", reply.Harbor)

	time.Sleep(200 * time.Second)
}
