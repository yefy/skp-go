package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "skp-go/skynet_go/logger"
	"testing"
	"time"
)

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

func Test_Gob1(t *testing.T) {
	log.SetLevel(log.Lerr)
	in := 10
	out := 0
	b := IntToBytes(in)
	out = BytesToInt(b)
	if in != out {
		t.Error()
	}
}

func Test_Gob2(t *testing.T) {

	ch := make(chan string)

	go func() {
		for m := range ch {
			fmt.Println("Processed:", m)
			time.Sleep(10 * time.Second) // 模拟需要长时间运行的操作
		}
	}()

	ch <- "cmd.1"
	ch <- "cmd.2" // 不会被接收处理
}
