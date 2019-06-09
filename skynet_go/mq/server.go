package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"skp-go/skynet_go/rpcdp"
	"skp-go/skynet_go/utility"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewSQProducer(server *Server, topic string, tag string) *SQProducer {
	q := &SQProducer{}
	q.server = server
	q.topic = topic
	q.tag = tag
	q.key = q.topic + "_" + q.tag
	q.SHProducer = NewSHProducer(nil)
	q.SHProducerInterface = q
	return q
}

type SQProducer struct {
	server *Server
	topic  string
	tag    string
	key    string
	*SHProducer
}

func (q *SQProducer) GetConn() bool {
	topicConns := q.server.topicConnsMap[q.topic]
	if topicConns != nil {
		for _, v := range topicConns.harborConn {
			if v.IsSubscribe(q.tag) {
				q.conn = v
				return true
			}
		}
	}

	q.conn = nil
	return false
}

func (q *SQProducer) GetTcpConn() bool {
	if q.conn == nil {
		if !q.GetConn() {
			return false
		}
	}

	if (q.conn.GetState() & connStateStopSubscribe) > 0 {
		q.conn = nil
		q.tcpConn = nil
		return false
	}

	return q.SHProducer.GetTcpConn()
}

func (q *SQProducer) Error() {
	q.SHProducer.Error()
	q.conn = nil
}

//==================SHConsumer================
func NewSHConsumer(conn *Conn) *SHConsumer {
	c := &SHConsumer{}
	c.conn = conn
	rpc.NewServer(c)
	return c
}

type SHConsumer struct {
	rpc.ServerBase
	conn       *Conn
	vector     *Vector
	tcpConn    *net.TCPConn
	tcpVersion int32
}

func (c *SHConsumer) Start() {
	c.Stop()
	c.RPC_GetServer().Start(false)
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("OnRead")
}

func (c *SHConsumer) Stop() {
	c.RPC_GetServer().Stop(true)
}

func (c *SHConsumer) GetTcpConn() bool {
	if c.tcpConn == nil && (c.conn.GetState()&connStateStart) > 0 {
		tcpConn, tcpVersion := c.conn.GetTcpConn()
		if tcpVersion != c.tcpVersion {
			c.tcpVersion = tcpVersion
			c.vector = NewVector()
		}
		c.tcpConn = tcpConn
		c.vector.SetConn(c.tcpConn)
	}

	if c.tcpConn != nil {
		return true
	}
	return false
}

func (c *SHConsumer) OnRead() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		if !c.GetTcpConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		mqMsg, err := c.ReadMqMsg(3)
		if err != nil {
			errCode := err.(*errorCode.ErrCode)
			if errCode.Code() != errorCode.TimeOut {
				c.tcpConn = nil
				c.conn.SendError(c.tcpVersion)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		if mqMsg == nil {
			continue
		}

		if mqMsg.GetTyp() == typRespond {
			harbor := mqMsg.GetHarbor()
			harborConnI, ok := c.conn.server.harborConn.Load(harbor)
			if ok {
				harborConn := harborConnI.(*Conn)
				harborConn.shProducer.SendWriteMqMsg(mqMsg)
			} else {
				//这里需要保存mqMsg用于排查问题
				log.Fatal("not harbor = %d", harbor)
			}
		} else {
			key := mqMsg.GetTopic() + "_" + mqMsg.GetTag()
			q := c.conn.server.topicTag[key]
			if q == nil {
				log.Fatal("NewQueue: mqMsg.GetTopic() = %+v, mqMsg.GetTag() = %+v", mqMsg.GetTopic(), mqMsg.GetTag())
				q = NewSQProducer(c.conn.server, mqMsg.GetTopic(), mqMsg.GetTag())
				c.conn.server.topicTag[key] = q
			}
			q.SendWriteMqMsg(mqMsg)
		}
	}
}

func (c *SHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil, nil
		}

		rMqMsg, err := c.getMqMsg()
		if err != nil {
			return nil, err
		}

		if rMqMsg != nil {
			return rMqMsg, nil
		}

		if err := c.vector.read(timeout); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (c *SHConsumer) getMqMsgSize() (int, error) {
	sizeBytes := c.vector.Get(4)
	if sizeBytes == nil {
		return 0, nil
	}

	size, err := utility.BytesToInt(sizeBytes)
	if err != nil {
		return 0, err
	}

	if !c.vector.checkSize(int(4 + size)) {
		return 0, nil
	}

	c.vector.Skip(4)
	return int(size), nil
}

func (c *SHConsumer) getMqMsg() (*MqMsg, error) {
	size, err := c.getMqMsgSize()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}
	msgByte := c.vector.Get(size)
	if msgByte == nil {
		return nil, nil
	}

	msg := &MqMsg{}
	if err := proto.Unmarshal(msgByte, msg); err != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	c.vector.Skip(size)
	log.Fatal("msg = %+v", proto.MarshalTextString(msg))
	return msg, nil
}

//=====================SHProducer===============

func NewSHProducer(conn *Conn) *SHProducer {
	p := &SHProducer{}
	p.conn = conn
	p.SHProducerInterface = p
	rpcdp.NewServer(p)
	return p
}

type SHProducerInterface interface {
	GetTcpConn() bool
	Error()
}

type SHProducer struct {
	rpcdp.ServerBase
	SHProducerInterface
	conn       *Conn
	tcpConn    *net.TCPConn
	tcpVersion int32
}

func (p *SHProducer) Start() {
	p.Stop()
	p.RPC_GetServer().Start(false)
}

func (p *SHProducer) Stop() {
	p.RPC_GetServer().Stop(false)
}

func (p *SHProducer) GetTcpConn() bool {
	if p.tcpConn == nil && (p.conn.GetState()&connStateStart) > 0 {
		tcpConn, tcpVersion := p.conn.GetTcpConn()
		p.tcpVersion = tcpVersion
		p.tcpConn = tcpConn
	}

	if p.tcpConn != nil {
		return true
	}
	return false
}

func (p *SHProducer) RPC_Dispath(method string, args []interface{}) error {
	if method == "OnWriteMqMsg" {
		mqMsg := args[0].(*MqMsg)
		return p.OnWriteMqMsg(mqMsg)
	}

	return nil
}

func (p *SHProducer) Error() {
	p.conn.SendError(p.tcpVersion)
	p.tcpConn = nil
}

func (p *SHProducer) SendWriteMqMsg(m *MqMsg) {
	p.RPC_GetServer().Send("OnWriteMqMsg", m)
}

func (p *SHProducer) OnWriteMqMsg(mqMsg *MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	for {
		if p.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil
		}

		if !p.SHProducerInterface.GetTcpConn() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := p.Write(mqMsgBytes); err != nil {
			p.SHProducerInterface.Error()
			continue
		}
		break
	}

	return nil
}

func (p *SHProducer) WriteMqMsg(mqMsg *MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))
	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if !p.SHProducerInterface.GetTcpConn() {
		log.Panic(errorCode.NewErrCode(0, "not GetTcpConn"))
	}
	log.Fatal("mqMsgBytes size = %d", len(mqMsgBytes))
	if err := p.Write(mqMsgBytes); err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) WriteSize(size int) error {
	var err error
	bytes, err := utility.IntToBytes(size)
	if err != nil {
		return err
	}

	err = p.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) Write(bytes []byte) error {
	if err := p.WriteSize(len(bytes)); err != nil {
		return err
	}

	err := p.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *SHProducer) WriteBytes(bytes []byte) error {
	size := len(bytes)
	for size > 0 {
		wSize, err := p.tcpConn.Write(bytes)
		if err != nil {
			return err
		}

		if wSize == size {
			break
		}
		bytes = bytes[wSize:]
		size -= wSize
	}
	return nil
}

//======================Conn================
func NewConn(server *Server, tcpConn *net.TCPConn) *Conn {
	c := &Conn{}
	c.server = server
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)
	c.SetTcpConn(tcpConn)
	rpc.NewServer(c)

	return c
}

type Conn struct {
	rpc.ServerBase
	server     *Server
	harbor     int32
	instance   string //topic_$$
	mutex      sync.Mutex
	tcpConn    *net.TCPConn
	tcpVersion int32
	state      int32

	topic  string
	tag    string
	tags   []string
	tagMap map[string]bool

	shConsumer *SHConsumer
	shProducer *SHProducer
}

func (c *Conn) SendError(tcpVersion int32) {
	c.RPC_GetServer().Send("OnError", tcpVersion)
}

func (c *Conn) OnError(tcpVersion int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if tcpVersion != c.tcpVersion {
		return
	}
	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
	c.tcpVersion++
	c.state = connStateErr
}

func (c *Conn) Start() {
	c.shConsumer.Start()
	c.shProducer.Start()
}

func (c *Conn) Stop() {
	c.shConsumer.Stop()
	c.shProducer.Stop()
}

func (c *Conn) Close() {
	c.Stop()
	c.RPC_GetServer().Stop(false)

	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}
}

func (c *Conn) SendOnCloseAll() {
	c.RPC_GetServer().Send("OnCloseAll")
}

func (c *Conn) OnCloseAll() {
	c.Close()
}

func (c *Conn) Subscribe(topic string, tag string) {
	c.topic = strings.Trim(topic, "\t\n ")
	c.tag = strings.Trim(tag, "\t\n ")
	if c.tag == "*" {
		return
	}

	c.tags = strings.Split(c.tag, "||")
	for _, v := range c.tags {
		c.tagMap[v] = true
	}
}

func (c *Conn) IsSubscribe(tag string) bool {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	log.Fatal("IsSubscribe tag = %s", tag)
	if c.tag == "*" {
		return true
	}

	if c.tagMap[tag] {
		return true
	}

	return false
}

func (c *Conn) IsSubscribeAll() bool {
	if c.tag == "*" {
		return true
	}

	return false
}

func (c *Conn) GetState() int32 {
	return c.state
}

func (c *Conn) GetTcpConn() (*net.TCPConn, int32) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn, c.tcpVersion
}

func (c *Conn) SetTcpConn(tcpConn *net.TCPConn) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
	}

	c.tcpConn = tcpConn
	c.tcpVersion++
	c.state = connStateStart
}

//==================TopicConns===========

type TopicConns struct {
	harborConn map[int32]*Conn
}

func NewTopicConns() *TopicConns {
	t := &TopicConns{}
	t.harborConn = make(map[int32]*Conn)
	return t
}

//==================Server===========

func NewServer() *Server {
	s := &Server{}
	rpc.NewServer(s)
	//s.instanceConn = make(map[string]*Conn)
	//	s.harborConn = make(map[int32]*Conn)
	s.topicConnsMap = make(map[string]*TopicConns)
	s.topicTag = make(map[string]*SQProducer)
	return s
}

type Server struct {
	rpc.ServerBase
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
	if newConn.shConsumer.GetTcpConn() == false {
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

	request := RegisteRequest{}
	if err := DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), &request); err != nil {
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
		reply := RegisterReply{}
		reply.Harbor = replyConn.harbor
		sMqMsg, err := ReplyMqMsg(replyConn.harbor, rMqMsg.GetPendingSeq(), rMqMsg.GetEncode(), &reply)
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
			newConn.tcpConn = nil
			newConn.Close()
			conn.SetTcpConn(tcpConn)
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
