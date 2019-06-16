package client

import (
	"fmt"
	"net"
	"os"
	"skp-go/skynet_go/encodes"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/rpc"
	"skp-go/skynet_go/rpc/rpcE"
	"skp-go/skynet_go/rpc/rpcU"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

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

func NewClient(instance string, address string) *Client {
	c := &Client{}
	rpcU.NewServer(c)
	c.address = address
	c.pendingSeq = 0
	c.instance = c.GetInstance(instance)
	c.pendingMsgPool = &sync.Pool{New: func() interface{} {
		msg := &PendingMsg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}
	c.dialConnI = conn.NewDialTcpConn(address)

	tcpConn, err := c.dialConnI.Connect()
	if err != nil {
		return nil
	}

	c.Client = mq.NewClient(tcpConn)

	c.chProducer = NewCHProducer(c)
	c.chConsumer = NewCHConsumer(c)
	c.rpcEMap = make(map[string]*rpcE.Server)
	return c
}

type ServerI interface {
	OnRegisterLocal(tcpConn conn.ConnI)
}

func NewLocalClient(instance string, server ServerI) *Client {
	c := &Client{}
	rpcU.NewServer(c)
	c.pendingSeq = 0
	c.instance = c.GetInstance(instance)
	c.pendingMsgPool = &sync.Pool{New: func() interface{} {
		msg := &PendingMsg{}
		msg.pending = make(chan interface{}, 1)
		return msg
	},
	}
	c.dialConnI = conn.NewDialMqConn()
	tcpConn, err := c.dialConnI.Connect()
	if err != nil {
		return nil
	}
	c.Client = mq.NewClient(tcpConn)

	server.OnRegisterLocal(c.dialConnI.GetS())

	c.chProducer = NewCHProducer(c)
	c.chConsumer = NewCHConsumer(c)
	c.rpcEMap = make(map[string]*rpcE.Server)
	return c
}

type Client struct {
	rpcU.ServerB
	address        string
	harbor         int32
	instance       string //xx_ip_$$ (模块名)_(ip)_(进程id)
	pendingSeq     uint64
	pendingMap     sync.Map
	pendingMsgPool *sync.Pool
	chProducer     *CHProducer
	chConsumer     *CHConsumer
	topic          string
	tag            string
	rpcEMap        map[string]*rpcE.Server
	*mq.Client
	dialConnI conn.DialConnI
}

func (c *Client) RegisterServer(obj rpcE.ServerI) {
	rpcServer := rpcE.NewServer(obj)
	c.rpcEMap[rpcServer.ObjectName()] = rpcServer
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

func (c *Client) DialConn() (conn.ConnI, error) {
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp4", c.address)
	if tcpAddrErr != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, tcpAddrErr.Error()))
	}

	tcpConn, tcpConnErr := net.DialTCP("tcp", nil, tcpAddr)
	if tcpConnErr != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, tcpConnErr.Error()))
	}

	return tcpConn, nil
}

func (c *Client) Register() error {
	request := mq.RegisteRequest{}
	request.Instance = c.instance
	request.Harbor = c.harbor
	request.Topic = c.topic
	request.Tag = c.tag

	reply := mq.RegisterReply{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq.OnClientRegister", &request, &reply); err != nil {
		return errorCode.NewErrCode(0, err.Error())
	}
	c.harbor = reply.Harbor

	log.Fatal("c.harbor = %d", c.harbor)
	return nil
}

func (c *Client) StopSubscribe() {
	request := mq.NilStruct{}
	reply := mq.NilStruct{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq.OnClientStopSubscribe", &request, &reply); err != nil {
		log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
	}

	c.Client.SetState(mq.ClientStateStopSubscribe)

	log.Fatal("StopSubscribe")
}

func (c *Client) Close() {
	c.Client.SetState(mq.ClientStateStopping)

	request := mq.NilStruct{}
	reply := mq.NilStruct{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq.OnClientClose", &request, &reply); err != nil {
		log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
	}

	c.Client.SetState(mq.ClientStateStop)

	log.Fatal("Close")
}

func (c *Client) WaitPending() bool {
	isExist := true
	c.pendingMap.Range(func(k, v interface{}) bool {
		log.Fatal("k= %+v, v = %+v", k, v)
		isExist = false
		return false
	})

	return isExist
}

func (c *Client) Exit() {
	//c.StopSubscribe()
	w := rpc.NewWait()
	w.Timer(c.WaitPending)
	//c.Close()
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
	p.typ = mq.TypeCall
	p.topic = msg.Topic
	p.tag = msg.Tag
	p.order = msg.Order
	p.class = methods[0]
	p.method = methods[1]
	p.pendingSeq = atomic.AddUint64(&c.pendingSeq, 1)
	p.reqEncode = msg.encode

	log.Fatal("request = %+v", request)
	var err error
	p.reqBody, err = encodes.EncodeBody(encodes.EncodeGob, request)
	if err != nil {
		return err
	}

	c.pendingMap.Store(p.pendingSeq, p)

	sMqMsg := &mq.MqMsg{}
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

	c.chProducer.SendWriteMqMsg(sMqMsg)
	//c.chProducer.RPC_GetServer().Send("SendMqMsg", sMqMsg)
	rMqMsgI := <-p.pending
	rMqMsg := rMqMsgI.(*mq.MqMsg)

	err = encodes.DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), reply)
	if err != nil {
		return err
	}
	log.Fatal("reply = %+v", reply)

	return nil
}
