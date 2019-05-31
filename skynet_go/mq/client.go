package mq

import (
	"fmt"
	"net"
	"os"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync"
	_ "sync/atomic"
	"time"
)

//==================HarborConsumer================
type CHConsumer struct {
	rpcServer *rpc.Server
	conn      *Conn
	tcpConn   *net.TCPConn
	vector    Vector
}

func NewCHConsumer(conn *Conn) *CHConsumer {
	c := &CHConsumer{}
	c.rpcServer = rpc.NewServer(c)
	c.conn = conn
	c.tcpConn = c.conn.GetTcpConn()
	c.vector.SetConn(c.tcpConn)
	return c
}

func (c *CHConsumer) Start() {
	c.rpcServer.Addoroutine(1)
	c.rpcServer.Send("Read")
}

func (c *CHConsumer) Read() {
	for {
		if c.rpcServer.IsStopping() {
			log.Fatal("rpcRead stop")
			return
		}

		msg, err := c.ReadMsg(5)
		if err != nil {
			continue
		}
		_ = msg

	}
}

func (c *CHConsumer) ReadMsg(timeout time.Duration) (*Msg, error) {
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

//=====================HarborProducer===============
type CHProducer struct {
	rpcServer *rpc.Server
	conn      *Conn
	tcpConn   *net.TCPConn
}

func NewCHProducer(conn *Conn) *CHProducer {
	p := &CHProducer{}
	p.rpcServer = rpc.NewServer(p)
	p.conn = conn
	p.tcpConn = p.conn.GetTcpConn()
	//sendList
	//waitList
	return p
}

func (c *CHProducer) Start() {
	c.rpcServer.Addoroutine(1)
	c.rpcServer.Send("Write")
}

func (p *CHProducer) SendMsg() {
}

func (p *CHProducer) Send() {
}

func (p *CHProducer) Write() {
	//return c.conn.Write(b)
}

const (
	Gob = iota
	Proto
)

type Msg struct {
	Topic string //模块名
	Tag   string //分标识
	Order uint64 //有序消息
	Code  int
}

type Client struct {
	rpcServer *rpc.Server
	address   string
	mutex     sync.Mutex
	tcpConn   *net.TCPConn
	harbor    int32
	instance  string //xx_ip_$$ (模块名)_(ip)_(进程id)

}

func NewClient(instance string, address string) *Client {
	c := &Client{}
	c.rpcServer = rpc.NewServer(c)
	c.address = address
	pid := os.Getpid()
	//addr := strings.Split(c.tcpConn.LocalAddr().String(), ":")[0]
	addr := c.getIP()
	c.instance = fmt.Sprintf("%s_%s_%d", instance, addr, pid)
	log.Fatal("instance = %s", c.instance)
	c.getConn()
	return c
}

func (c *Client) getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
		return ""
	}
	//log.Fatal("len(addrs) = %d", len(addrs))
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//log.Fatal("ip = %s", ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (c *Client) getConn() {
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp4", c.address)
	if tcpAddrErr != nil {
		log.Panic(errorCode.NewErrCode(0, tcpAddrErr.Error()))
		return
	}
	tcpConn, tcpConnErr := net.DialTCP("tcp", nil, tcpAddr)
	if tcpConnErr != nil {
		log.ErrorCode(errorCode.NewErrCode(0, tcpConnErr.Error()))
		return
	}

	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.tcpConn = tcpConn
}

func (c *Client) Register() {

}

func (c *Client) Send(msg *Msg, method string, args interface{}) error {
	return nil
}

func (c *Client) SendReq(msg *Msg, method string, args interface{}, reply interface{}, replyFunc interface{}) error {
	return nil
}

func (c *Client) Call(msg *Msg, method string, args interface{}, reply interface{}) error {
	return nil
}
