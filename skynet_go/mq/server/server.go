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

	s.RPC_GetServer().Addoroutine(1)
	if err := s.RPC_GetServer().Send("Accept"); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	return nil
}

func (s *Server) Accept() {
	for {
		tcpConn, err := s.listen.AcceptTCP()
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		log.Fatal("本地IP地址: %s, 远程IP地址:%s", tcpConn.LocalAddr(), tcpConn.RemoteAddr())
		//输出：220.181.111.188:80
		go s.OnRegister(tcpConn)
	}
}

func (s *Server) Close() {
	if err := s.listen.Close(); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	s.RPC_GetServer().Stop(true)
}

func (s *Server) OnRegister(tcpConn *net.TCPConn) {
	var err error
	newClient := NewClient(s, tcpConn)
	if newClient.shConsumer.Consumer.GetTcp() == false {
		log.Err("GetTcp error")
		return
	}

	rMqMsg, err := newClient.shConsumer.ReadMqMsg(3)
	if err != nil {
		newClient.Close()
		return
	}

	if rMqMsg.GetClass() != "Mq" ||
		rMqMsg.GetMethod() != "OnRegister" {
		log.Err("rMqMsg.Class != Mq || rMqMsg.Method != Register , rMqMsg = %+v", rMqMsg)
		newClient.Close()
		return
	}

	request := mq.RegisteRequest{}
	if err := encodes.DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), &request); err != nil {
		newClient.Close()
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
			replyClient.shProducer.WriteMqMsg(sMqMsg)
		}
	}

	if request.Harbor > 0 {
		connI, connOk := s.instanceClient.Load(request.Instance)
		if connOk == false {
			log.Err("not request.Instance = %+v", request.Instance)
			newClient.Close()
			return
		}

		conn := connI.(*Client)
		if conn.harbor != request.Harbor {
			log.Err("conn.harbor != request.Harbor, conn.harbor = %+v, request.Harbor = %+v", conn.harbor, request.Harbor)
			newClient.Close()
		} else {
			replyFunc(newClient)
			newClient.Client.ClearTcp()
			newClient.Close()
			conn.SetTcp(tcpConn)
		}
	} else {
		newClient.instance = request.Instance
		newClient.harbor = atomic.AddInt32(&s.harbor, 1)
		newClient.Subscribe(request.Topic, request.Tag)

		_, connOk := s.instanceClient.LoadOrStore(request.Instance, newClient)
		if connOk {
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
					conn.SendOnCloseAll()
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
							conn.SendOnCloseAll()
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
		newClient.Start()
	}
}
