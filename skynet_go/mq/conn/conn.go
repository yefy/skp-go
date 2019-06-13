package conn

import (
	"time"
)

func NewConnChan() *ConnChan {
	c := &ConnChan{}
	c.Cache = make(chan []byte, 1000)
	return c
}

type ConnChan struct {
	cc    *ConnChan
	Cache chan []byte
}

func (c *ConnChan) SetReadDeadline(t time.Time) error {
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
	rb := <-c.cc.Cache
	copy(b, rb)
	return len(b), nil
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
