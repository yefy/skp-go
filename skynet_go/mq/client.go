package mq

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

//==================HarborConsumer================
func NewCHConsumer(client *Client) *CHConsumer {
	c := &CHConsumer{}
	rpc.NewServer(c)
	c.c = client
	c.tcpConn = c.c.GetTcpConn()
	c.vector = NewVector()
	c.vector.SetConn(c.tcpConn)
	c.Start()
	return c
}

type CHConsumer struct {
	rpc.ServerBase
	c       *Client
	tcpConn *net.TCPConn
	vector  *Vector
}

func (c *CHConsumer) Start() {
	c.RPC_GetServer().Addoroutine(1)
	c.RPC_GetServer().Send("Read")
}

func (c *CHConsumer) Read() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		rMqMsg, err := c.ReadMqMsg(0)
		if err != nil {
			return
		}

		if rMqMsg.GetTyp() == typRespond {
			pendingMsgI, ok := c.c.pendingMap.Load(rMqMsg.GetPendingSeq())
			if !ok {
				log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
				continue
			}
			pendingMsg := pendingMsgI.(*PendingMsg)
			if pendingMsg.typ == typCall {
				pendingMsg.pending <- rMqMsg
			}
		}
	}
}

func (c *CHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
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
		log.Fatal("c.vector.read(timeout) timeout = %d", timeout)
		if err := c.vector.read(timeout); err != nil {
			return nil, errorCode.NewErrCode(0, err.Error())
		}
	}

	return nil, nil
}

func (c *CHConsumer) getMqMsgSize() int {
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

func (c *CHConsumer) getMqMsg() (*MqMsg, error) {
	size := c.getMqMsgSize()
	log.Fatal("CHConsumer size = %d", size)
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

/*
func (c *CHConsumer) Read() {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return
		}

		rMqMsg, err := c.ReadMqMsg(0)
		if err != nil {
			continue
		}
		if rMqMsg.Ty

		sMqMsg := &MqMsg{}
	sMqMsg.Typ = proto.Int32(typRespond)
	sMqMsg.Harbor = proto.Int32(conn.harbor)
	sMqMsg.PendingSeq = proto.Uint64(rMqMsg.PendingSeq)
	sMqMsg.Encode = proto.Int32(encodeGob)
	sMqMsg.Body = proto.String("")
	conn.shConsumer.SendMqMsg(sMqMsg)

	}
}

func (c *CHConsumer) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		if c.RPC_GetServer().IsStop() {
			log.Fatal("rpcRead stop")
			return nil, nil
		}
		//获取msg 如果有返回  如果没有接收数据 超时返回错误
		c.vector.read(timeout)
	}

	return nil, nil
}
*/
//=====================NewCHProducer===============
func NewCHProducer(c *Client) *CHProducer {
	p := &CHProducer{}
	rpc.NewServer(p)
	p.c = c
	p.GetTcpConn()
	return p
}

type CHProducer struct {
	rpc.ServerBase
	c       *Client
	tcpConn *net.TCPConn
}

func (p *CHProducer) GetTcpConn() {
	if p.tcpConn == nil {
		p.tcpConn = p.c.GetTcpConn()
	}
}

func (p *CHProducer) SendMqMsg(m *MqMsg) {
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

}

func (p *CHProducer) WriteSize(size int) {
	log.Fatal("size = %d", size)
	s := int32(size)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, s)
	p.tcpConn.Write(bytesBuffer.Bytes())
}

func (p *CHProducer) Write(b []byte) {
	p.WriteSize(len(b))
	p.tcpConn.Write(b)
}

const (
	Gob = iota
	Proto
)

//=====================Msg

type Msg struct {
	Topic  string //模块名
	Tag    string //分标识
	Order  uint64 //有序消息
	encode int32
}

type CallBack func(error)
type PendingMsg struct {
	typ        int32  //Send SendReq Call CallReq
	topic      string //模块名
	tag        string //分标识
	order      uint64 //有序消息
	class      string //远端对象名字
	method     string //远端对象的方法
	pendingSeq uint64 //回调seq
	reqEncode  int32  //编码 gob  proto
	reqBody    string //数据包
	resEncode  int32  //编码 gob  proto
	resBody    string //数据包
	callBack   CallBack
	pending    chan interface{}
}

func (p *PendingMsg) Init() {

}

//=====================Client
func NewClient(instance string, address string) *Client {
	c := &Client{}
	rpc.NewServer(c)
	c.address = address
	c.pendingSeq = 0
	c.instance = c.GetInstance(instance)
	c.pendingMsgPool = &sync.Pool{New: func() interface{} {
		msg := &PendingMsg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}

	if err := c.DialConn(); err != nil {
		return nil
	}

	c.chProducer = NewCHProducer(c)
	c.chConsumer = NewCHConsumer(c)
	return c
}

type Client struct {
	rpc.ServerBase
	address        string
	mutex          sync.Mutex
	tcpConn        *net.TCPConn
	harbor         int32
	instance       string //xx_ip_$$ (模块名)_(ip)_(进程id)
	pendingSeq     uint64
	pendingMap     sync.Map
	pendingMsgPool *sync.Pool
	chProducer     *CHProducer
	chConsumer     *CHConsumer
	topic          string
	tag            string
}

func (c *Client) Subscribe(topic string, tag string) {
	c.topic = topic
	c.tag = tag
}

func (c *Client) Start() error {
	if err := c.Register(); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetInstance(instance string) string {
	if len(c.instance) < 1 {
		pid := os.Getpid()
		//addr := strings.Split(c.tcpConn.LocalAddr().String(), ":")[0]
		addr := c.getIP()
		c.instance = fmt.Sprintf("%s_%s_%d", instance, addr, pid)
		log.Fatal("instance = %s", c.instance)
	}
	return c.instance
}

func (c *Client) getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
		return ""
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (c *Client) DialConn() error {
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp4", c.address)
	if tcpAddrErr != nil {
		return log.Panic(errorCode.NewErrCode(0, tcpAddrErr.Error()))
	}

	tcpConn, tcpConnErr := net.DialTCP("tcp", nil, tcpAddr)
	if tcpConnErr != nil {
		return log.Panic(errorCode.NewErrCode(0, tcpConnErr.Error()))
	}

	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.tcpConn = tcpConn
	return nil
}

func (c *Client) GetTcpConn() *net.TCPConn {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.tcpConn
}

func (c *Client) Register() error {
	request := RegisteRequest{}
	request.Instance = c.instance
	request.Harbor = c.harbor
	request.Topic = c.topic
	request.Tag = c.tag
	reply := RegisterReply{}
	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq.OnRegister", &request, &reply); err != nil {
		return errorCode.NewErrCode(0, err.Error())
	}
	c.harbor = reply.Harbor

	log.Fatal("c.harbor = %d", c.harbor)
	return nil
}

// func (c *Client) Send(msg *Msg, method string, request interface{}) error {
// 	return nil
// }

// func (c *Client) SendReq(msg *Msg, method string, request interface{}, reply interface{}, callBack CallBack) {

// 	callBack(errorCode.NewErrCode(0, "test"))
// 	return nil
// }

func (c *Client) Call(msg *Msg, method string, request interface{}, reply interface{}) error {
	methods := strings.Split(method, ".")
	p := c.pendingMsgPool.Get().(*PendingMsg)
	p.typ = typCall
	p.topic = msg.Topic
	p.tag = msg.Tag
	p.order = msg.Order
	p.class = methods[0]
	p.method = methods[1]
	p.pendingSeq = atomic.AddUint64(&c.pendingSeq, 1)
	p.reqEncode = msg.encode

	log.Fatal("request = %+v", request)

	var sBuf bytes.Buffer
	enc := gob.NewEncoder(&sBuf)
	enc.Encode(request)
	p.reqBody = sBuf.String()

	c.pendingMap.Store(p.pendingSeq, p)

	sMqMsg := &MqMsg{}
	sMqMsg.Typ = proto.Int32(p.typ)
	sMqMsg.Harbor = proto.Int32(c.harbor)
	sMqMsg.Topic = proto.String(p.topic)
	sMqMsg.Tag = proto.String(p.tag)
	sMqMsg.Order = proto.Uint64(p.order)
	sMqMsg.Class = proto.String(p.class)
	sMqMsg.Method = proto.String(p.method)
	sMqMsg.PendingSeq = proto.Uint64(p.pendingSeq)
	sMqMsg.Encode = proto.Int32(p.reqEncode)
	sMqMsg.Body = proto.String(p.reqBody)

	c.chProducer.RPC_GetServer().Send("SendMqMsg", sMqMsg)
	rMqMsgI := <-p.pending
	rMqMsg := rMqMsgI.(*MqMsg)

	var rBuf bytes.Buffer
	rBuf.WriteString(rMqMsg.GetBody())
	dec := gob.NewDecoder(&rBuf)
	dec.Decode(reply)

	log.Fatal("reply = %+v", reply)

	// mqMsgByte, err := proto.Marshal(&sMqMsg)
	// if err != nil {
	// 	return log.Panic(errorCode.NewErrCode(0, err.Error()))
	// }
	// if true {
	// 	mqMsg2 := MqMsg{}
	// 	proto.Unmarshal(mqMsgByte, &mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", mqMsg2)
	// 	log.Fatal("mqMsg2 = %+v", proto.MarshalTextString(&mqMsg2))

	// }
	return nil
}
