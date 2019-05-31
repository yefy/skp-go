package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync"
	"sync/atomic"
	"time"
)

type MqMsg struct {
	Typ        string //Send SendReq Call CallReq
	Harbor     int32  //harbor 全局唯一的id  对应instance
	Instance   string //xx_ip_$$ (模块名)_(ip)_(进程id)
	Topic      string //模块名
	Tag        string //分标识
	Order      uint64 //有序消息
	Class      string //远端对象名字
	Method     string //远端对象的方法
	PendingSeq uint64 //回调seq
	BodyType   string //编码 gob  proto
	Body       string //数据包
}

func NewMqMsg() *MqMsg {
	return nil
}

type Queue struct {
	list     string
	backList string
	conn     *Conn
}

func NewQueue() *Queue {
	return nil
}

func (q *Queue) Write() {

}

//==================SHConsumer================
type SHConsumer struct {
	rpcServer *rpc.Server
	conn      *Conn
	tcpConn   *net.TCPConn
	vector    Vector
}

func NewSHConsumer(conn *Conn) *SHConsumer {
	c := &SHConsumer{}
	c.rpcServer = rpc.NewServer(c)
	c.conn = conn
	c.tcpConn = c.conn.GetTcpConn()
	c.vector.SetConn(c.tcpConn)
	return c
}

func (c *SHConsumer) Start() {
	c.rpcServer.Addoroutine(1)
	c.rpcServer.Send("Read")
}

func (c *SHConsumer) Read() {
	for {
		if c.rpcServer.IsStopping() {
			log.Fatal("rpcRead stop")
			return
		}

		msg, err := c.ReadMqMsg(5)
		if err != nil {
			continue
		}
		_ = msg

	}
}

func (c *SHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		if c.rpcServer.IsStopping() {
			log.Fatal("rpcRead stop")
			return nil, nil
		}
		//获取msg 如果有返回  如果没有接收数据 超时返回错误
		c.vector.read(timeout)
	}

	return nil, nil
}

//=====================SHProducer===============
type SHProducer struct {
	rpcServer *rpc.Server
	conn      *Conn
	tcpConn   *net.TCPConn
}

func NewSHProducer(conn *Conn) *SHProducer {
	p := &SHProducer{}
	p.rpcServer = rpc.NewServer(p)
	p.conn = conn
	p.tcpConn = p.conn.GetTcpConn()
	//sendList
	//waitList
	return p
}

func (c *SHProducer) Start() {
	c.rpcServer.Addoroutine(1)
	c.rpcServer.Send("Write")
}

func (p *SHProducer) SendMqMsg() {
}

func (p *SHProducer) Send() {
}

func (p *SHProducer) Write() {
	//return c.conn.Write(b)
}

//======================Conn================
type Conn struct {
	server   *Server
	harbor   int32
	instance string //topic_$$
	mutex    sync.Mutex
	tcpConn  *net.TCPConn

	topic    string
	tag      string
	topicTag map[string]bool

	shConsumer *SHConsumer
	shProducer *SHProducer
}

func NewConn(server *Server, tcpConn *net.TCPConn, harbor int32) *Conn {
	c := &Conn{}
	c.server = server
	c.harbor = harbor
	c.tcpConn = tcpConn
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)
	return c
}

func (c *Conn) GetTcpConn() *net.TCPConn {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn
}

func (c *Conn) SetTcpConn(tcpconn *net.TCPConn) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.tcpConn = tcpconn
}

func (c *Conn) Start() {
	c.shConsumer.Start()
	c.shProducer.Start()
}

// func (c *Conn) Stop() {
// 	c.conn.Close()
// 	c.rpcRead.Stop(false)
// }

// func (c *Conn) IsTopicTag() {
// }

type TopicConns struct {
	harborConn map[int32]*Conn
}

func NewTopicConns() *TopicConns {
	t := &TopicConns{}
	t.harborConn = make(map[int32]*Conn)
	return t
}

//==================Server===========
type Server struct {
	rpcServer *rpc.Server
	listen    *net.TCPListener
	harbor    int32
	//instanceConn map[string]*Conn
	instanceConn sync.Map
	//harborConn   map[int32]*Conn
	harborConn sync.Map
	topicConns map[string]*TopicConns
}

func NewServer() *Server {
	s := &Server{}
	s.rpcServer = rpc.NewServer(s)
	return s
}

func (s *Server) Listen(address string) error {
	var tcpaddr *net.TCPAddr
	var err error
	if tcpaddr, err = net.ResolveTCPAddr("tcp4", address); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if s.listen, err = net.ListenTCP("tcp", tcpaddr); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	s.rpcServer.Addoroutine(1)
	if err := s.rpcServer.Send("Accept"); err != nil {
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

		conn := NewConn(s, tcpConn, atomic.AddInt32(&s.harbor, 1))

		go func(newConn *Conn) {
			msg, msgErr := newConn.shConsumer.ReadMqMsg(2)
			if msgErr != nil {
				newConn.tcpConn.Close()
				return
			}
			if msg.Class != "Mq" ||
				msg.Method != "Register" ||
				len(msg.Instance) < 1 {
				log.Err("msg.Class != Mq || msg.Method != Register, msg = %+v", msg)
				newConn.tcpConn.Close()
				return
			}
			newConn.instance = msg.Instance

			connI, connOk := s.instanceConn.LoadOrStore(newConn.instance, newConn)
			if connOk {
				oldConn := connI.(Conn)
				if msg.Harbor != oldConn.harbor {
					log.Err("msg.Harbor != oldConn.Harbor, msg.Harbor = %+v, oldConn.Harbor = %+v", msg.Harbor, oldConn.harbor)
					newConn.tcpConn.Close()
				} else {
					oldConn.SetTcpConn(newConn.tcpConn)
				}
			} else {
				s.harborConn.Store(newConn.harbor, newConn)
				newConn.Start()
			}
		}(conn)
	}
}

func (s *Server) Close() {
	if err := s.listen.Close(); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	s.rpcServer.Stop(true)
}
