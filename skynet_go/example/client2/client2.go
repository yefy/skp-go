package main

import (
	"fmt"
	"runtime"
	"skp-go/skynet_go/defaultConf"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/client"
	"skp-go/skynet_go/rpc/rpcE"
	"time"
)

//taskkill /im server.test.exe /f
//taskkill /im client.test.exe /f
//go test server_test.go server.go mqMsg.pb.go vector.go

type Client2Test struct {
	rpcE.ServerB
}

func (c *Client2Test) RPC_Describe() string {
	return "Client2Test"
}

func (c *Client2Test) Client2TestHello(in *mq.RegisteRequest, out *mq.RegisterReply) error {
	log.Fatal("Client2Test in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func NewClient(index int) *client.Client {
	key := fmt.Sprintf("Client2Test_%d", index)
	mqClient := client.NewClient(key, ":5673")
	mqClient.Subscribe(key, "*")
	mqClient.RegisterServer(&Client2Test{})
	mqClient.Start()

	return mqClient
}

func RunClient(index int) {
	var harbor int32 = 1000000000
	mqClient := NewClient(index)
	time.Sleep(time.Second * 5)
	var timeout int = 0
	for {
		request := mq.RegisteRequest{}
		request.Instance = "Instance"
		harbor++
		request.Harbor = harbor
		request.Topic = "Topic"
		request.Tag = "Tag"

		reply := mq.RegisterReply{}

		msg := client.Msg{Topic: "MqTest", Tag: "0"}
		time1 := time.Now().UnixNano() / 1e6
		if err := mqClient.Call(&msg, "MqTest.OnMqTestHello", &request, &reply); err != nil {
			log.Fatal("error")
		}
		time2 := time.Now().UnixNano() / 1e6
		diff := time2 - time1
		//log.Fatal("diff = %d, time1 = %d, time2 = %d", diff, time1, time2)
		if diff > 1000 {
			log.Fatal("diff = %d > 1", diff)
			timeout++
		}
		log.Fatal("reply.Harbor = %d, timeout = %d", reply.Harbor, timeout)

		if harbor != reply.Harbor {
			panic(reply.Harbor)
		}
	}
	mqClient.Close()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defaultConf.SetDebug()
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Ltrace, log.Lscreen)
	log.Fatal("client1 start NumCPU = %d", runtime.NumCPU())

	for i := 0; i < 100; i++ {
		index := i
		go RunClient(index)
	}
	wait := make(chan bool)
	<-wait
}
