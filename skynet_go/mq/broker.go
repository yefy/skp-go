package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/rpc"
	"sync"
	_ "sync/atomic"
	"time"
)

type Conn struct {
	instance *Instance
}

func (c *conn)Read(){
	
}

func (c *conn)Write(){
	
}

func (c *)

type Instance struct {
	session int32
	conn    *net.TCPConn
	topic string
	tag string
}

type Queue struct {
	list string
	conns []*conn
}



type Tag struct {
	tag map[string] *Queue
}


type Broker struct {
	rpcBroker   *rpc.Server
	listen      *net.TCPListener
	session     int32
	mutex       *sync.Mutex
	instanceMap map[string]*Instance
	
	topic map[string]*Tag
}

func NewBroker() *Broker {
	broker := &Broker{}
	broker.rpcBroker = rpc.NewServer(broker)
	return broker
}

func (b *Broker) Listen(address string) error {
	var tcpaddr *net.TCPAddr
	var err error
	if tcpaddr, err = net.ResolveTCPAddr("tcp4", address); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if b.listen, err = net.ListenTCP("tcp", tcpaddr); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	b.rpcBroker.Addoroutine(1)
	if err := b.rpcBroker.Send("Accept"); err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	return nil
}

func (b *Broker) Accept() {
	for {
		conn, err := b.listen.AcceptTCP()
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		go func(conn *net.TCPConn) {
			//读取对象名字
			var instanceName string
			b.mutex.Lock()
			instance := b.instanceMap[instanceName]
			if instance == nil {
				b.session++
				instance = &Instance{}
				instance.session = b.session
				instance.conn = conn
				b.instanceMap[instanceName] = instance
			} else {
				if instance.conn {
					instance.conn.Close()
				}
				instance.conn = conn
			}
			b.mutex.Unlock()
			// 返回  session

		}(conn)

		// b.rpcBroker.Addoroutine(1)
		// if err := b.rpcBroker.Send("Conn", conn); err != nil {
		// 	log.Panic(errorCode.NewErrCode(0, err.Error()))
		// }
	}
}

func (b *Broker) Conn(conn *net.TCPConn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	v := NewVector()
	for {
		if b.rpcBroker.IsStop() {
			log.Fatal("rpcBroker stop")
			return
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		size, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err.Error())
			continue
		}
		v.Put(buf[:size])
		log.Fatal("buf = %s", string(v.GetAll()))
	}
}

func (b *Broker) Close() {
	if err := b.listen.Close(); err != nil {
		log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	b.rpcBroker.Stop()
}
