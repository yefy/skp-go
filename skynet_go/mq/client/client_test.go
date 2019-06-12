package client

import (
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcEncode"
	"testing"
)

type Test struct {
	rpcEncode.ServerBase
}

func (t *Test) OnRegister(in *mq.RegisteRequest, out *mq.RegisterReply) error {
	log.Fatal("in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func Test_Client1(t *testing.T) {
	log.SetLevel(log.Lerr)
	mqClient := NewClient("Test", ":5673")
	mqClient.Subscribe("Test", "*")
	mqClient.RegisterServer(&Test{})
	mqClient.Start()

	request := mq.RegisteRequest{}
	request.Instance = "Instance"
	request.Harbor = 13
	request.Topic = "Topic"
	request.Tag = "Tag"
	reply := mq.RegisterReply{}
	msg := Msg{Topic: "Test", Tag: "0"}
	if err := mqClient.Call(&msg, "Test.OnRegister", &request, &reply); err != nil {
		t.Error()
	}

	log.Fatal("reply.Harbor = %d", reply.Harbor)
}
