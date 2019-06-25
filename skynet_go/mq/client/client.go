package client

import (
	"fmt"
	"net"
	"os"
	"skp-go/skynet_go/encodes"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"skp-go/skynet_go/rpc/rpcE"
	"skp-go/skynet_go/rpc/rpcU"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	c.dialConnI = mq.NewDialTcpConn(address)

	connI, err := c.dialConnI.Connect()
	if err != nil {
		return nil
	}

	c.Client = mq.NewClient(connI)

	c.chProducer = NewCHProducer(c)
	c.chConsumer = NewCHConsumer(c)

	for i := 0; i < len(c.rpcEMapArr); i++ {
		c.rpcEMapArr[i] = make(map[string]*rpcE.Server)
	}
	return c
}

type ServerI interface {
	AddLocalClient(connI mq.ConnI)
}

func NewLocalClient(instance string, serverI ServerI) *Client {
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
	c.dialConnI = mq.NewDialMqConn()
	connI, err := c.dialConnI.Connect()
	if err != nil {
		return nil
	}

	c.Client = mq.NewClient(connI)
	serverI.AddLocalClient(c.dialConnI.GetS())

	c.chProducer = NewCHProducer(c)
	c.chConsumer = NewCHConsumer(c)

	for i := 0; i < len(c.rpcEMapArr); i++ {
		c.rpcEMapArr[i] = make(map[string]*rpcE.Server)
	}
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
	rpcEMapArr     [10]map[string]*rpcE.Server
	*mq.Client
	dialConnI mq.DialConnI
}

func (c *Client) RegisterServer(obj rpcE.ServerI) {
	rpcServer := rpcE.NewServer(obj)
	for i := 0; i < len(c.rpcEMapArr); i++ {
		rpcEMap := c.rpcEMapArr[i]
		rpcEMap[rpcServer.ObjectName()] = rpcServer
	}
}

func (c *Client) Subscribe(topic string, tag string) {
	c.topic = topic
	c.tag = tag
}

func (c *Client) Start() error {
	if err := c.MqRegister(); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetDescribe() string {
	return c.instance
}

func (c *Client) GetPendingMsg(rMqMsg *mq.MqMsg) *PendingMsg {
	pendingMsgI, ok := c.pendingMap.Load(rMqMsg.GetPendingSeq())
	if !ok {
		log.Fatal("not rMqMsg.PendingSeq = %d", rMqMsg.PendingSeq)
		return nil
	}
	c.pendingMap.Delete(rMqMsg.GetPendingSeq())
	pendingMsg := pendingMsgI.(*PendingMsg)
	return pendingMsg
}

func (c *Client) GetRPCServer(rMqMsg *mq.MqMsg) *rpcE.Server {
	var rpcEMap map[string]*rpcE.Server
	order := rMqMsg.GetOrder()
	if order == 0 {
		order = 0
	} else if order%uint64(len(c.rpcEMapArr)) == 0 {
		order = 1
	} else {
		order = order % uint64(len(c.rpcEMapArr))
	}

	rpcEMap = c.rpcEMapArr[order]

	rpcServer := rpcEMap[rMqMsg.GetClass()]
	if rpcServer == nil {
		log.Fatal("not rMqMsg.GetClass() = %+v", rMqMsg.GetClass())
		return nil
	}
	return rpcServer
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

// func (c *Client) DialConn() (mq.ConnI, error) {
// 	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp4", c.address)
// 	if tcpAddrErr != nil {
// 		return nil, log.Panic(errorCode.NewErrCode(0, tcpAddrErr.Error()))
// 	}

// 	tcpConn, tcpConnErr := net.DialTCP("tcp", nil, tcpAddr)
// 	if tcpConnErr != nil {
// 		return nil, log.Panic(errorCode.NewErrCode(0, tcpConnErr.Error()))
// 	}

// 	return tcpConn, nil
// }

func (c *Client) RegisterMqMsg(mqMsg *mq.MqMsg) {
	if mqMsg.GetClass() != "Mq" {
		log.Err("mqMsg.GetClass() != Mq")
		return
	}

	if mqMsg.GetMethod() == mq.OnMqClosing {
		go c.MqStopSubscribe()
	} else {
		log.Err("not class = %+v or method = %+v", mqMsg.GetClass(), mqMsg.GetMethod())
	}
}

func (c *Client) MqRegister() error {
	request := mq.RegisteRequest{}
	request.Instance = c.instance
	request.Harbor = c.harbor
	request.Topic = c.topic
	request.Tag = c.tag

	reply := mq.RegisterReply{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq."+mq.OnMqRegister, &request, &reply); err != nil {
		return errorCode.NewErrCode(0, err.Error())
	}
	c.harbor = reply.Harbor

	log.Fatal("c.harbor = %d", c.harbor)
	return nil
}

func (c *Client) MqStopSubscribe() {
	request := mq.NilStruct{}
	reply := mq.NilStruct{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq."+mq.OnMqStopSubscribe, &request, &reply); err != nil {
		log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
	}

	c.Client.SetState(mq.ClientStateStopSubscribe)

	log.Fatal("MqStopSubscribe")
}

func (c *Client) MqClose() {
	c.Client.SetState(mq.ClientStateStopping)

	request := mq.NilStruct{}
	reply := mq.NilStruct{}

	msg := Msg{Topic: "Mq", Tag: "*"}
	if err := c.Call(&msg, "Mq."+mq.OnMqClose, &request, &reply); err != nil {
		log.ErrorCode(errorCode.NewErrCode(0, err.Error()))
	}

	c.Client.SetState(mq.ClientStateStop)

	log.Fatal("MqClose")
}

func (c *Client) WaitPending() bool {
	c.pendingMap.Range(func(k, v interface{}) bool {
		log.Fatal("k= %+v, v = %+v", k, v)
		return false
	})

	return true
}

func (c *Client) Close() {
	if true {
		log.Fatal("Client Close")
		return
	}

	c.MqStopSubscribe()
	c.RPC_GetServer().Ticker(time.Second, c.WaitPending)

	isTimeout := c.RPC_GetServer().Timer(time.Second*10, c.MqClose)
	if isTimeout {
		log.Err("MqClose timeout")
	}
	c.chProducer.Stop()
	c.chConsumer.Stop()
	c.Client.CloseConn()
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

	c.chProducer.RpcSend_OnWriteMqMsg(sMqMsg)

	rMqMsgI := <-p.pending
	rMqMsg := rMqMsgI.(*mq.MqMsg)

	err = encodes.DecodeBody(rMqMsg.GetEncode(), rMqMsg.GetBody(), reply)
	if err != nil {
		return err
	}
	log.Fatal("reply = %+v", reply)

	return nil
}
