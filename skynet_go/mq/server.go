package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync"
	_ "sync/atomic"
	_ "time"
)

type Msg struct {
	typ        string
	harbor     int32
	topic      string
	tag        string
	order      uint64
	instance   string
	class      string
	method     string
	pendingSeq uint64
	bodyType   string //gob  proto
	body       string
}

type Queue struct {
	list     string
	backList string
	conn     *Conn
}

func (q *Queue) Write() {

}

type Conn struct {
	server      *Server
	harbor      int32
	instance    string //topic_$$
	conn        *net.TCPConn
	topic       string
	tag         string
	harborQueue *Queue
}

func (i *Conn) Read() {

}

func (i *Conn) Write() {

}

type Server struct {
	rpcServer     *rpc.Server
	listen        *net.TCPListener
	mutex         *sync.Mutex
	harbor        int32
	instanceConn  map[string]*Conn
	topicTagConn  map[string]*Conn
	harborQueue   map[int32]*Queue
	topicTagQueue map[string]*Queue
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
		_ = conn
		// go func(conn *net.TCPConn) {
		// 	//读取对象名字
		// 	var instanceName string
		// 	s.mutex.Lock()
		// 	instance := s.instanceMap[instanceName]
		// 	if instance == nil {
		// 		s.session++
		// 		instance = &Instance{}
		// 		instance.session = s.session
		// 		instance.conn = conn
		// 		s.instanceMap[instanceName] = instance
		// 	} else {
		// 		if instance.conn {
		// 			instance.conn.Close()
		// 		}
		// 		instance.conn = conn
		// 	}
		// 	s.mutex.Unlock()
		// 	// 返回  session

		// }(conn)

		// s.rpcServer.Addoroutine(1)
		// if err := s.rpcServer.Send("Conn", conn); err != nil {
		// 	log.Panic(errorCode.NewErrCode(0, err.Error()))
		// }
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
	s.rpcServer.Stop()
}
