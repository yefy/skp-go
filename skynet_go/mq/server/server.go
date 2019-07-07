package server

import (
	"net"
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
	mutex        sync.Mutex
	harborClient map[int32]*Client
}

func (t *TopicClients) IsOk(c *Client) bool {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	if c.IsSubscribeAll() && len(t.harborClient) > 0 {
		log.Err("tag exist")
		return false
	}

	for _, client := range t.harborClient {
		for _, tag := range c.tags {
			if client.IsSubscribe(tag) {
				log.Err("tag exist")
				return false
			}
		}
	}

	t.harborClient[c.harbor] = c

	return true
}

func (t *TopicClients) GetClient(tag string) *Client {
	defer t.mutex.Unlock()
	t.mutex.Lock()

	for _, v := range t.harborClient {
		if v.IsSubscribe(tag) {
			return v
		}
	}
	return nil
}

func (t *TopicClients) Del(c *Client) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	delete(t.harborClient, c.harbor)
}

func NewServer() *Server {
	s := &Server{}
	rpcU.NewServer(s)
	return s
}

type Server struct {
	rpcU.ServerB
	listen         *net.TCPListener
	harbor         int32
	instanceClient sync.Map
	harborClient   sync.Map
	topicClients   sync.Map
	topicTag       sync.Map
	waitClient     sync.Map
}

func (s *Server) RPC_Describe() string {
	return "mq.Server"
}

func (s *Server) GetDescribe() string {
	return "mq.Server"
}

func (s *Server) GetHarbor() int32 {
	return atomic.AddInt32(&s.harbor, 1)
}

func (s *Server) Listen(address string, isWait bool) error {
	var err error
	tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	s.listen, err = net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	if isWait {
		s.OnAccept()
	} else {
		s.rpcSend_OnAccept()
	}
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

		//s.listen.SetDeadline(time.Now().Add(3 * time.Second))
		tcpConn, err := s.listen.AcceptTCP()
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			continue
		}

		if err != nil {
			log.Fatal(err.Error())
			return
		}

		log.Debug("本地IP地址: %s, 远程IP地址:%s", tcpConn.LocalAddr(), tcpConn.RemoteAddr())
		go s.AddClient(tcpConn)
	}
}

func (s *Server) AddClient(connI mq.ConnI) {
	c := NewClient(s, connI)
	c.Start()
	s.waitClient.Store(c, c)
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

func (s *Server) GetClient(topic string, tag string) *Client {
	if topicClientsI, ok := s.topicClients.Load(topic); ok {
		topicClients := topicClientsI.(*TopicClients)
		return topicClients.GetClient(tag)
	}

	return nil
}

func (s *Server) GetHarborClient(mqMsg *mq.MqMsg) *Client {
	harbor := mqMsg.GetHarbor()
	if harborClientI, ok := s.harborClient.Load(harbor); ok {
		return harborClientI.(*Client)
	}
	//这里需要保存mqMsg用于排查问题
	log.Fatal("not harbor = %d", harbor)
	return nil
}

func (s *Server) GetSQProducer(mqMsg *mq.MqMsg) *SQProducer {
	key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
	if qI, ok := s.topicTag.Load(key); ok {
		return qI.(*SQProducer)
	}

	q := NewSQProducer(s, mqMsg.GetTopic(), mqMsg.GetTag())
	if qI, ok := s.topicTag.LoadOrStore(key, q); ok {
		q.Stop()
		return qI.(*SQProducer)
	}
	log.Debug("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
	return q
}

func (s *Server) OnMqDel(c *Client) {
	s.instanceClient.Delete(c.instance)
	s.harborClient.Delete(c.harbor)
	if topicClientsI, ok := s.topicClients.Load(c.topic); ok {
		topicClients := topicClientsI.(*TopicClients)
		topicClients.Del(c)
	}
}

func (s *Server) OnMqRegister(c *Client) bool {
	s.waitClient.Delete(c)

	if _, ok := s.instanceClient.LoadOrStore(c.instance, c); ok {
		log.Err("c.instance exist")
		return false
	}

	s.harborClient.Store(c.harbor, c)

	var topicClients *TopicClients
	topicClientsI, ok := s.topicClients.Load(c.topic)
	if !ok {
		topicClients = NewTopicClients()
		if topicClientsI, ok := s.topicClients.LoadOrStore(c.topic, topicClients); ok {
			topicClients = topicClientsI.(*TopicClients)
		}
	} else {
		topicClients = topicClientsI.(*TopicClients)
	}
	if !topicClients.IsOk(c) {
		return false
	}

	return true
}
