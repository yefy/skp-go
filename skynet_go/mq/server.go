package mq

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"skp-go/skynet_go/rpcdp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

// type MqMsg struct {
// 	Typ        string //Send SendReq Call CallReq
// 	Harbor     int32  //harbor 全局唯一的id  对应instance
// 	Instance   string //xx_ip_$$ (模块名)_(ip)_(进程id)
// 	Topic      string //模块名
// 	Tag        string //分标识
// 	Order      uint64 //有序消息
// 	Class      string //远端对象名字
// 	Method     string //远端对象的方法
// 	PendingSeq uint64 //回调seq
// 	BodyType   string //编码 gob  proto
// 	Body       string //数据包
// }

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
func NewSHConsumer(conn *Conn) *SHConsumer {
	c := &SHConsumer{}
	rpc.NewServer(c)
	c.conn = conn
	c.tcpConn = c.conn.GetTcpConn()
	c.vector = NewVector()
	c.vector.SetConn(c.tcpConn)
	return c
}

type SHConsumer struct {
	rpc.ServerBase
	conn    *Conn
	tcpConn *net.TCPConn
	vector  *Vector
}

func (c *SHConsumer) Start() {
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("Read")
}

func (c *SHConsumer) Read() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		msg, err := c.ReadMqMsg(0)
		if err != nil {
			continue
		}
		_ = msg

	}
}

func (c *SHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	n := 0
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

		//获取msg 如果有返回  如果没有接收数据 超时返回错误
		if err := c.vector.read(timeout); err != nil {
			return nil, errorCode.NewErrCode(0, err.Error())
		}
		n++
		if n > 3 {
			return nil, nil
		}
	}

	return nil, nil
}

// //整形转换成字节
// func IntToBytes(n int) []byte {
//     x := int32(n)

//     bytesBuffer := bytes.NewBuffer([]byte{})
//     binary.Write(bytesBuffer, binary.BigEndian, x)
//     return bytesBuffer.Bytes()
// }

// //字节转换成整形
// func BytesToInt(b []byte) int {
//     bytesBuffer := bytes.NewBuffer(b)

//     var x int32
//     binary.Read(bytesBuffer, binary.BigEndian, &x)

//     return int(x)
// }

func (c *SHConsumer) getMqMsgSize() int {
	sizeByte := c.vector.Get(4)
	if sizeByte == nil {
		return 0
	}
	bytesBuffer := bytes.NewBuffer(sizeByte)
	var size int32
	if err := binary.Read(bytesBuffer, binary.BigEndian, &size); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
		return 0
	}

	if !c.vector.checkSize(int(4 + size)) {
		return 0
	}

	c.vector.Skip(4)
	return int(size)
}

func (c *SHConsumer) getMqMsg() (*MqMsg, error) {
	size := c.getMqMsgSize()
	log.Fatal("size = %d", size)
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
	rpcdp.NewServer(p)
	p.conn = conn
	p.tcpConn = p.conn.GetTcpConn()
	return p
}

type SHProducer struct {
	rpcdp.ServerBase
	conn    *Conn
	tcpConn *net.TCPConn
}

func (p *SHProducer) RPC_Dispath(method string, args []interface{}) error {
	m := args[0].(*MqMsg)
	log.Fatal("msg = %+v", proto.MarshalTextString(m))
	mb, err := proto.Marshal(m)
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	// if false {
	// 	mqMsg2 := MqMsg{}
	// 	proto.Unmarshal(mqMsgByte, &mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", proto.MarshalTextString(&mqMsg2))

	// }
	p.Write(mb)
	return nil
}

func (p *SHProducer) SendMqMsg(m *MqMsg) {
	p.RPC_GetServer().Send("Write", m)
}

func (p *SHProducer) WriteSize(size int) {
	log.Fatal("size = %d", size)
	s := int32(size)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, s)
	p.tcpConn.Write(bytesBuffer.Bytes())
}

func (p *SHProducer) Write(b []byte) {
	p.WriteSize(len(b))
	p.tcpConn.Write(b)
}

//======================Conn================
func NewConn(server *Server, tcpConn *net.TCPConn, harbor int32) *Conn {
	c := &Conn{}
	c.server = server
	c.harbor = harbor
	c.tcpConn = tcpConn
	c.shConsumer = NewSHConsumer(c)
	c.shProducer = NewSHProducer(c)
	return c
}

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

func NewServer() *Server {
	s := &Server{}
	rpc.NewServer(s)
	s.topicConnsMap = make(map[string]*TopicConns)
	return s
}

type Server struct {
	rpc.ServerBase
	listen *net.TCPListener
	harbor int32
	//instanceConn map[string]*Conn
	instanceConn sync.Map
	//harborConn   map[int32]*Conn
	harborConn    sync.Map
	topicConnsMap map[string]*TopicConns
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
	newConn := NewConn(s, tcpConn, atomic.AddInt32(&s.harbor, 1))
	rMqMsg, rMqMsgErr := newConn.shConsumer.ReadMqMsg(2)
	if rMqMsgErr != nil {
		newConn.tcpConn.Close()
		return
	}

	request := RegisteRequest{}
	var rBuf bytes.Buffer
	rBuf.WriteString(rMqMsg.GetBody())
	dec := gob.NewDecoder(&rBuf)
	dec.Decode(&request)

	log.Fatal("request = %+v", request)

	if rMqMsg.GetClass() != "Mq" ||
		rMqMsg.GetMethod() != "OnRegister" ||
		request.Instance == "" {
		log.Err("rMqMsg.Class != Mq || rMqMsg.Method != Register, rMqMsg = %+v", rMqMsg)
		newConn.tcpConn.Close()
		return
	}
	newConn.instance = request.Instance
	if request.Harbor > 0 {
		newConn.harbor = request.Harbor
	}

	connI, connOk := s.instanceConn.LoadOrStore(newConn.instance, newConn)
	log.Fatal("connI = %+v, connOk = %+v", connI, connOk)
	conn := connI.(*Conn)
	if connOk {
		if newConn.harbor != conn.harbor {
			log.Err("newConn.harbor != conn.Harbor, newConn.harbor = %+v, oldConn.Harbor = %+v", newConn.harbor, conn.harbor)
			conn.tcpConn.Close()
		} else {
			conn.SetTcpConn(newConn.tcpConn)
		}
	} else {
		s.harborConn.Store(newConn.harbor, newConn)
		//conn.Start()
	}

	topicConns := s.topicConnsMap[conn.topic]
	if topicConns == nil {
		topicConns = NewTopicConns()
		topicConns.harborConn[conn.harbor] = conn
		s.topicConnsMap[conn.topic] = topicConns
	} else {

	}

	reply := RegisterReply{}
	reply.Harbor = conn.harbor

	var sBuf bytes.Buffer
	enc := gob.NewEncoder(&sBuf)
	enc.Encode(&reply)

	sMqMsg := &MqMsg{}
	sMqMsg.Typ = proto.Int32(typRespond)
	sMqMsg.Harbor = proto.Int32(conn.harbor)
	sMqMsg.PendingSeq = proto.Uint64(rMqMsg.GetPendingSeq())
	sMqMsg.Encode = proto.Int32(encodeGob)
	sMqMsg.Body = proto.String(sBuf.String())
	sMqMsg.Topic = proto.String("")
	sMqMsg.Tag = proto.String("")
	sMqMsg.Order = proto.Uint64(0)
	sMqMsg.Class = proto.String("")
	sMqMsg.Method = proto.String("")
	conn.shProducer.SendMqMsg(sMqMsg)
}
