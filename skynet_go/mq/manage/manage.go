package manage

import (
	"skp-go/skynet_go/mq/client"
	"skp-go/skynet_go/mq/server"
	"skp-go/skynet_go/rpc/rpcE"
)

type Mq struct {
	rpcE.ServerB
}

func (t *Test) OnRegister(in *mq.RegisteRequest, out *mq.RegisterReply) error {
	log.Fatal("in = %+v", in)
	out.Harbor = in.Harbor
	return nil
}

func NewManage(s *server.Server) *Manage {
	m := &Manage{}
	m.c = client.NewLocalClient(s)
	mqClient.Subscribe("Mq", "*")
	mqClient.RegisterServer(&Test{})
	mqClient.Start()
	return m
}

type Manage struct {
	c *client.Client
}
