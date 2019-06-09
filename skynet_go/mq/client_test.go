package mq

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq/rpcGob"
	"testing"
)

type Test struct {
	rpcGob.ServerBase
}

func (t *Test) OnRegister(in *RegisteRequest, out *RegisterReply) error {
	log.Fatal("in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func Test_Client1(t *testing.T) {
	log.SetLevel(log.Lerr)
	mqClient := NewClient("Test", ":5672")
	mqClient.Subscribe("Test", "*")
	mqClient.RegisterServer(&Test{})
	mqClient.Start()

	request := RegisteRequest{}
	request.Instance = "Instance"
	request.Harbor = 13
	request.Topic = "Topic"
	request.Tag = "Tag"
	reply := RegisterReply{}
	msg := Msg{Topic: "Test", Tag: "0"}
	if err := mqClient.Call(&msg, "Test.OnRegister", &request, &reply); err != nil {
		t.Error()
	}

	log.Fatal("reply.Harbor = %d", reply.Harbor)
}
