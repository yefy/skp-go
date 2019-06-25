package server

import (
	"net"
	"skp-go/skynet_go/encodes"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcU"
	"sync"
	"sync/atomic"
	"time"
)

func NewTopicClients() *TopicClients {
	t := &TopicClients{}
	t.harborClient = make(map[int32]*Client)
	return t
}

type TopicClients struct {
	harborClient map[int32]*Client
}

func NewServer() *Server {
	s := &Server{}
	rpcU.NewServer(s)
	//s.instanceClient = make(map[string]*Client)
	//	s.harborClient = make(map[int32]*Client)
	s.topicClientsMap = make(map[string]*TopicClients)
	s.topicTag = make(map[string]*SQProducer)
	return s
}

type Server struct {
	rpcU.ServerB
	listen *net.TCPListener
	harbor int32
	mutex  sync.Mutex
	//instanceClient map[string]*Client
	instanceClient sync.Map
	//harborClient map[int32]*Client
	harborClient    sync.Map
	topicClientsMap map[string]*TopicClients
	topicTag        map[string]*SQProducer
}

func (s *Server) Listen(address string) error {
	var err error
	tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	s.listen, err = net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	s.rpcSend_OnAccept()
	return nil
}

func (s *Server) rpcSend_OnAccept() {
	s.RPC_GetServer().Addoroutine(1)
	s.RPC_GetServer().Send("OnAccept")
}

func (s *Server) OnAccept() {
	for {
		if s.RPC_GetServer().IsStop() {
			log.Fatal("Accept stop")
			return
		}

		s.listen.SetDeadline(time.Now().Add(3 * time.Second))
		tcpConn, err := s.listen.AcceptTCP()
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			continue
		}

		if err != nil {
			log.Fatal(err.Error())
			return
		}

		log.Fatal("本地IP地址: %s, 远程IP地址:%s", tcpConn.LocalAddr(), tcpConn.RemoteAddr())
		go s.AddClient(tcpConn)
	}
}

func (s *Server) AddClient(connI mq.ConnI) {
	newClient := NewClient(s, connI)
	newClient.Start()
}

func (s *Server) AddLocalClient(connI mq.ConnI) {
	go s.AddClient(connI)
}

func (s *Server) Close() {
	if err := s.listen.Close(); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	s.RPC_GetServer().Stop(true)
}

func (s *Server) OnClientRegister(c *Client, rMqMsg *mq.MqMsg) {
	newClient := c
	request := mq.RegisteRequest{}
	if err := encodes.DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), &request); err != nil {
		c.Close()
		return
	}
	log.Fatal("request = %+v", request)

	if request.Instance == "" {
		log.Err("not request.Instance")
		newClient.Close()
		return
	}

	replyFunc := func(replyClient *Client) {
		reply := mq.RegisterReply{}
		reply.Harbor = replyClient.harbor
		sMqMsg, err := mq.ReplyMqMsg(replyClient.harbor, rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), &reply)
		if err == nil {
			replyClient.shProducer.RpcSend_OnWriteMqMsg(sMqMsg)
		}
	}

	if request.Harbor > 0 {
		clientI, clientOk := s.instanceClient.Load(request.Instance)
		if clientOk == false {
			log.Err("not request.Instance = %+v", request.Instance)
			newClient.Close()
			return
		}

		client := clientI.(*Client)
		if client.harbor != request.Harbor {
			log.Err("client.harbor != request.Harbor, client.harbor = %+v, request.Harbor = %+v", client.harbor, request.Harbor)
			newClient.Close()
		} else {
			replyFunc(newClient)
			connI := newClient.Client.ClearConn2()
			newClient.Close()
			client.SetConn(connI)
		}
	} else {
		newClient.instance = request.Instance
		newClient.harbor = atomic.AddInt32(&s.harbor, 1)
		newClient.Subscribe(request.Topic, request.Tag)

		_, clientOk := s.instanceClient.LoadOrStore(request.Instance, newClient)
		if clientOk {
			log.Err("exist request.Instance = %+v", request.Instance)
			newClient.Close()
			return
		}
		log.Fatal("1111111111111111")
		replyFunc(newClient)
		topicClients := s.topicClientsMap[newClient.topic]
		if topicClients == nil {
			topicClients = NewTopicClients()
			s.topicClientsMap[newClient.topic] = topicClients
		} else {
			if newClient.IsSubscribeAll() {
				log.Fatal("IsSubscribeAll")
				//topicClientsMap all conn  close
				for harbor, conn := range topicClients.harborClient {
					log.Fatal("SendOnCloseAll harbor", harbor)
					conn.CloseSelf()
				}
				topicClients = NewTopicClients()
				s.topicClientsMap[newClient.topic] = topicClients
			} else {
				//关闭一样的tag conn
				var delHarbors []int32
				for harbor, conn := range topicClients.harborClient {
					for _, tag := range newClient.tags {
						if conn.IsSubscribe(tag) {
							log.Fatal("SendOnCloseAll harbor", harbor)
							conn.CloseSelf()
							delHarbors = append(delHarbors, harbor)
							break
						}
					}
				}

				for _, harbor := range delHarbors {
					delete(topicClients.harborClient, harbor)
				}
			}
		}

		topicClients.harborClient[newClient.harbor] = newClient

		s.harborClient.Store(newClient.harbor, newClient)
		//newClient.Start()
	}
}

func (s *Server) GetClient(topic string, tag string) *Client {
	topicClients := s.topicClientsMap[topic]
	if topicClients != nil {
		for _, v := range topicClients.harborClient {
			if v.IsSubscribe(tag) {
				return v
			}
		}
	}
	return nil
}

func (s *Server) GetHarborClient(mqMsg *mq.MqMsg) *Client {
	harbor := mqMsg.GetHarbor()
	harborClientI, ok := s.harborClient.Load(harbor)
	if ok {
		harborClient := harborClientI.(*Client)
		return harborClient
	}
	//这里需要保存mqMsg用于排查问题
	log.Fatal("not harbor = %d", harbor)
	return nil
}

func (s *Server) GetSQProducer(mqMsg *mq.MqMsg) *SQProducer {
	key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
	q := s.topicTag[key]
	if q == nil {
		log.Fatal("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
		q = NewSQProducer(s, mqMsg.GetTopic(), mqMsg.GetTag())
		s.topicTag[key] = q
	}
	return q
}
