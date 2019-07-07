package main

import (
	"skp-go/skynet_go/defaultConf"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/client"
	"skp-go/skynet_go/rpc/rpcE"
)

//taskkill /im server.test.exe /f
//taskkill /im client.test.exe /f
//go test server_test.go server.go mqMsg.pb.go vector.go

type Client1Test struct {
	rpcE.ServerB
}

func (c *Client1Test) RPC_Describe() string {
	return "Client1Test"
}

func (c *Client1Test) Client1TestHello(in *mq.RegisteRequest, out *mq.RegisterReply) error {
	log.Fatal("Client1Test in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func NewClient() *client.Client {
	log.Fatal("NewClient 000000000")
	mqClient := client.NewClient("Client1Test", ":5673")
	log.Fatal("Subscribe 000000000")
	mqClient.Subscribe("Client1Test", "*")
	log.Fatal("RegisterServer 000000000")
	mqClient.RegisterServer(&Client1Test{})
	log.Fatal("NewClient 11111111 ")
	mqClient.Start()
	log.Fatal("NewClient 22222222 ")
	return mqClient
}

func main() {
	defaultConf.SetDebug()
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Ltrace, log.Lscreen)
	log.Fatal("client1 start")

	var harbor int32 = 0

	for {
		mqClient := NewClient()

		request := mq.RegisteRequest{}
		request.Instance = "Instance"
		harbor++
		request.Harbor = harbor
		request.Topic = "Topic"
		request.Tag = "Tag"

		reply := mq.RegisterReply{}
		msg := client.Msg{Topic: "MqTest", Tag: "0"}
		log.Fatal("NewClient 333333 ")
		if err := mqClient.Call(&msg, "MqTest.OnMqTestHello", &request, &reply); err != nil {
			log.Fatal("error")
		}
		log.Fatal("NewClient 444444 ")
		log.Fatal("reply.Harbor = %d", reply.Harbor)

		if harbor != reply.Harbor {
			panic(reply.Harbor)
		}

		mqClient.Close()
	}
}
