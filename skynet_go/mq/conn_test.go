package mq

import (
	"fmt"
	"testing"
)

func Test_Conn(t *testing.T) {
	c := NewConn()
	ss := c.getS()
	cc := c.getC()
	go func() {
		b := make([]byte, 4096)
		ss.Read(b)
		fmt.Println(b)
		ss.Write([]byte("222"))
	}()
	cc.Write([]byte("111"))
	b := make([]byte, 4096)
	cc.Read(b)
	fmt.Println(b)
}
