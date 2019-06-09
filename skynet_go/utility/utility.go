package utility

import (
	"bytes"
	"encoding/binary"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
)

//整形转换成字节
func IntToBytes(n int) ([]byte, error) {
	x := int32(n)

	buffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(buffer, binary.BigEndian, x); err != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	return buffer.Bytes(), nil
}

//字节转换成整形
func BytesToInt(b []byte) (int, error) {
	buffer := bytes.NewBuffer(b)

	var x int32
	if err := binary.Read(buffer, binary.BigEndian, &x); err != nil {
		return 0, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	return int(x), nil
}
