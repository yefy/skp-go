package main

import (
	"runtime"
	"skp-go/skynet_go/defaultConf"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/client"
	"skp-go/skynet_go/mq/server"
	"skp-go/skynet_go/rpc/rpcE"
	"time"
)

//taskkill /im mq.exe /f
//taskkill /im client1.exe /f
//taskkill /im client2.exe /f

type MqTest struct {
	rpcE.ServerB
}

func (mt *MqTest) RPC_Describe() string {
	return "MqTest"
}

func (mt *MqTest) OnMqTestHello(in *mq.RegisteRequest, out *mq.RegisterReply) error {
	log.All("MqTestHello in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func LocalClientTest(s *server.Server) {
	time.Sleep(time.Second * 1)

	mqClient := client.NewLocalClient("MqTest", s)
	mqClient.Subscribe("MqTest", "*")
	mqClient.RegisterServer(&MqTest{})
	mqClient.Start()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defaultConf.SetDebug()
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Ltrace, log.Lscreen)
	log.Fatal("mq start NumCPU = %d", runtime.NumCPU())

	s := server.NewServer()
	go LocalClientTest(s)
	s.Listen(":5673", true)
}
