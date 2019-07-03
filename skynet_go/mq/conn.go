package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"time"
)

type ConnChanError struct {
}

func (c *ConnChanError) Error() string {
	return "ConnChan timeout"
}

func (c *ConnChanError) Timeout() bool {
	return true
}

func (c *ConnChanError) Temporary() bool {
	return true
}

func NewConnChan() *ConnChan {
	c := &ConnChan{}
	c.Cache = make(chan []byte, 1000)
	return c
}

type ConnChan struct {
	cc    *ConnChan
	Cache chan []byte
	t     time.Time
}

func (c *ConnChan) SetReadDeadline(t time.Time) error {
	c.t = t
	return nil
}

func (c *ConnChan) SetConnChan(cc *ConnChan) {
	c.cc = cc
}

func (c *ConnChan) Write(b []byte) (int, error) {
	c.Cache <- b
	return len(b), nil
}

func (c *ConnChan) Read(b []byte) (int, error) {
	// rb := <-c.cc.Cache
	// copy(b, rb)
	// return len(rb), nil
	t := c.t.Sub(time.Now())
	timer := time.NewTimer(t)
	for {
		select {
		case <-timer.C:
			var err net.Error = &ConnChanError{}
			return 0, err
		case rb := <-c.cc.Cache:
			copy(b, rb)
			return len(rb), nil
		}
	}
}

func (c *ConnChan) Close() error {
	return nil
}

type ConnI interface {
	SetReadDeadline(t time.Time) error
	Write(b []byte) (int, error)
	Read(b []byte) (int, error)
	Close() error
}

func NewConn() *Conn {
	c := &Conn{}
	c.c = NewConnChan()
	c.s = NewConnChan()
	c.c.SetConnChan(c.s)
	c.s.SetConnChan(c.c)
	return c
}

type Conn struct {
	c *ConnChan
	s *ConnChan
}

func (c *Conn) getC() ConnI {
	return c.c
}

func (c *Conn) getS() ConnI {
	return c.s
}

type DialConnI interface {
	Connect() (ConnI, error)
	GetS() ConnI
}

func NewDialMqConn() *DialMqConn {
	d := &DialMqConn{}
	return d
}

type DialMqConn struct {
	conn *Conn
}

func (d *DialMqConn) Connect() (ConnI, error) {
	d.conn = NewConn()
	return d.conn.getC(), nil
}

func (d *DialMqConn) GetS() ConnI {
	return d.conn.getS()
}

func NewDialTcpConn(address string) *DialTcpConn {
	d := &DialTcpConn{}
	d.address = address
	return d
}

type DialTcpConn struct {
	address string
	conn    *net.TCPConn
}

func (d *DialTcpConn) Connect() (ConnI, error) {
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp4", d.address)
	if tcpAddrErr != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, tcpAddrErr.Error()))
	}

	tcpConn, tcpConnErr := net.DialTCP("tcp", nil, tcpAddr)
	if tcpConnErr != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, tcpConnErr.Error()))
	}
	d.conn = tcpConn
	return tcpConn, nil
}

func (d *DialTcpConn) GetS() ConnI {
	return nil
}
