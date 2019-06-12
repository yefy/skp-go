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

func NewTopicConns() *TopicConns {
	t := &TopicConns{}
	t.harborConn = make(map[int32]*Conn)
	return t
}

type TopicConns struct {
	harborConn map[int32]*Conn
}

func NewServer() *Server {
	s := &Server{}
	rpcU.NewServer(s)
	//s.instanceConn = make(map[string]*Conn)
	//	s.harborConn = make(map[int32]*Conn)
	s.topicConnsMap = make(map[string]*TopicConns)
	s.topicTag = make(map[string]*SQProducer)
	return s
}

type Server struct {
	rpcU.ServerB
	listen *net.TCPListener
	harbor int32
	mutex  sync.Mutex
	//instanceConn map[string]*Conn
	instanceConn sync.Map
	//harborConn map[int32]*Conn
	harborConn    sync.Map
	topicConnsMap map[string]*TopicConns
	topicTag      map[string]*SQProducer
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
	newConn := NewConn(s, tcpConn)
	if newConn.shConsumer.Consumer.GetTcp() == false {
		log.Err("GetTcpConn error")
		return
	}

	rMqMsg, err := newConn.shConsumer.ReadMqMsg(3)
	if err != nil {
		newConn.Close()
		return
	}

	if rMqMsg.GetClass() != "Mq" ||
		rMqMsg.GetMethod() != "OnRegister" {
		log.Err("rMqMsg.Class != Mq || rMqMsg.Method != Register , rMqMsg = %+v", rMqMsg)
		newConn.Close()
		return
	}

	request := mq.RegisteRequest{}
	if err := encodes.DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), &request); err != nil {
		newConn.Close()
		return
	}
	log.Fatal("request = %+v", request)

	if request.Instance == "" {
		log.Err("not request.Instance")
		newConn.Close()
		return
	}

	replyFunc := func(replyConn *Conn) {
		reply := mq.RegisterReply{}
		reply.Harbor = replyConn.harbor
		sMqMsg, err := mq.ReplyMqMsg(replyConn.harbor, rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), &reply)
		if err == nil {
			replyConn.shProducer.WriteMqMsg(sMqMsg)
		}
	}

	if request.Harbor > 0 {
		connI, connOk := s.instanceConn.Load(request.Instance)
		if connOk == false {
			log.Err("not request.Instance = %+v", request.Instance)
			newConn.Close()
			return
		}

		conn := connI.(*Conn)
		if conn.harbor != request.Harbor {
			log.Err("conn.harbor != request.Harbor, conn.harbor = %+v, request.Harbor = %+v", conn.harbor, request.Harbor)
			newConn.Close()
		} else {
			replyFunc(newConn)
			newConn.Conn.ClearTcp()
			newConn.Close()
			conn.SetTcp(tcpConn)
		}
	} else {
		newConn.instance = request.Instance
		newConn.harbor = atomic.AddInt32(&s.harbor, 1)
		newConn.Subscribe(request.Topic, request.Tag)

		_, connOk := s.instanceConn.LoadOrStore(request.Instance, newConn)
		if connOk {
			log.Err("exist request.Instance = %+v", request.Instance)
			newConn.Close()
			return
		}
		log.Fatal("1111111111111111")
		replyFunc(newConn)
		topicConns := s.topicConnsMap[newConn.topic]
		if topicConns == nil {
			topicConns = NewTopicConns()
			s.topicConnsMap[newConn.topic] = topicConns
		} else {
			if newConn.IsSubscribeAll() {
				log.Fatal("IsSubscribeAll")
				//topicConnsMap all conn  close
				for harbor, conn := range topicConns.harborConn {
					log.Fatal("SendOnCloseAll harbor", harbor)
					conn.SendOnCloseAll()
				}
				topicConns = NewTopicConns()
				s.topicConnsMap[newConn.topic] = topicConns
			} else {
				//关闭一样的tag conn
				var delHarbors []int32
				for harbor, conn := range topicConns.harborConn {
					for _, tag := range newConn.tags {
						if conn.IsSubscribe(tag) {
							log.Fatal("SendOnCloseAll harbor", harbor)
							conn.SendOnCloseAll()
							delHarbors = append(delHarbors, harbor)
							break
						}
					}
				}

				for _, harbor := range delHarbors {
					delete(topicConns.harborConn, harbor)
				}
			}
		}

		topicConns.harborConn[newConn.harbor] = newConn

		s.harborConn.Store(newConn.harbor, newConn)
		newConn.Start()
	}
}
