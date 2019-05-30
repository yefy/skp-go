package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync"
	"sync/atomic"
	_ "time"
)

type Msg struct {
	Typ        string //Send SendReq Call CallReq
	Harbor     int32  //harbor 全局唯一的id  对应instance
	Instance   string //xx_$$ (模块名)_(进程id)
	Topic      string //模块名
	Tag        string //分标识
	Order      uint64 //有序消息
	Class      string //远端对象名字
	Method     string //远端对象的方法
	PendingSeq uint64 //回调seq
	BodyType   string //编码 gob  proto
	Body       string //数据包
}

func NewMsg() *Msg {
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

type Conn struct {
	vector   Vector
	server   *Server
	harbor   int32
	instance string //topic_$$
	conn     *net.TCPConn

	topic       string
	tag         string
	topicTag    map[string]bool
	harborQueue *Queue
}

func NewConn() *Conn {
	c := &Conn{}
	return c
}

func (i *Conn) Start() {
}

func (i *Conn) IsTopicTag() {
}

func (i *Conn) ReadMsg() (*Msg, error) {
	return nil, nil
}

func (i *Conn) Write() {

}

type TopicConns struct {
	harborConn map[int32]*Conn
}

func NewTopicConns() *TopicConns {
	t := &TopicConns{}
	t.harborConn = make(map[int32]*Conn)
	return t
}

type Server struct {
	rpcServer *rpc.Server
	listen    *net.TCPListener
	mutex     *sync.Mutex
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
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		c := NewConn()
		c.server = s
		c.conn = conn
		c.harbor = atomic.AddInt32(&s.harbor, 1)

		go func(newConn *Conn) {
			msg, msgErr := newConn.ReadMsg()
			if msgErr != nil {
				newConn.conn.Close()
				return
			}
			if msg.Class != "Mq" ||
				msg.Method != "Register" ||
				len(msg.Instance) < 1 {
				log.Err("msg.Class != Mq || msg.Method != Register, msg = %+v", msg)
				newConn.conn.Close()
				return
			}

			oldConnI, oldConnOk := s.instanceConn.LoadOrStore(msg.Instance, newConn)
			if oldConnOk {
				oldConn := oldConnI.(Conn)
				if msg.Harbor != oldConn.harbor {
					log.Err("msg.Harbor != oldConn.Harbor, msg.Harbor = %+v, oldConn.Harbor = %+v", msg.Harbor, oldConn.harbor)
					newConn.conn.Close()
				} else {
					oldConn.conn = newConn.conn
				}
			} else {
				newConn.instance = msg.Instance
				newConn.Start()
			}
		}(c)
	}
}

// func (s *Server) Conn(conn *net.TCPConn) {
// 	defer conn.Close()
// 	buf := make([]byte, 4096)
// 	v := NewVector()
// 	for {
// 		if s.rpcServer.IsStop() {
// 			log.Fatal("rpcServer stop")
// 			return
// 		}

// 		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
// 		size, err := conn.Read(buf)
// 		if err != nil {
// 			log.Fatal(err.Error())
// 			continue
// 		}
// 		v.Put(buf[:size])
// 		log.Fatal("buf = %s", string(v.GetAll()))
// 	}
// }

func (s *Server) Close() {
	if err := s.listen.Close(); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	s.rpcServer.Stop(true)
}
