package mq

import (
	"net"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	_ "strings"
	"time"
)

type Vector struct {
	buffer     []byte
	connBuffer []byte
	tcpConn    *net.TCPConn
}

func NewVector() *Vector {
	v := &Vector{}
	v.buffer = make([]byte, 0, 4096)
	v.connBuffer = make([]byte, 4096)
	return v
}

func (v *Vector) SetConn(tcpConn *net.TCPConn) {
	v.tcpConn = tcpConn
}

func (v *Vector) read(timeout time.Duration) error {
	if timeout > 0 {
		v.tcpConn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	}

	size, err := v.tcpConn.Read(v.connBuffer)
	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			log.Fatal("timeout")
			return errorCode.NewErrCode(errorCode.TimeOut, err.Error())
		}

		log.Fatal(err.Error())
		return errorCode.NewErrCode(errorCode.Unknown, err.Error())
	}

	if size > 0 {
		v.Put(v.connBuffer[:size])
	}
	return nil
}

func (v *Vector) Put(buf []byte) {
	v.buffer = append(v.buffer, buf...)
}
func (v *Vector) checkSize(size int) bool {
	buffLen := len(v.buffer)
	if size > buffLen {
		return false
	}
	return true
}

func (v *Vector) Get(size int) []byte {
	buffLen := len(v.buffer)
	if size > buffLen {
		return nil
	}
	return v.buffer[:size]
}

func (v *Vector) GetAll() []byte {
	buffLen := len(v.buffer)
	if buffLen <= 0 {
		return nil
	}
	return v.buffer[:buffLen]
}

func (v *Vector) Skip(size int) {
	buffLen := len(v.buffer)
	if size > buffLen {
		log.Panic(errorCode.NewErrCode(0, "size(%d) > buffLen(%d)", size, buffLen))
	}
	v.buffer = v.buffer[size:]
}

func (v *Vector) SkipAll() {
	buffLen := len(v.buffer)
	if buffLen > 0 {
		v.buffer = v.buffer[buffLen:]
	}
}
