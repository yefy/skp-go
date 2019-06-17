package mq

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq/conn"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewMqConn() *MqConn {
	mc := &MqConn{}
	return mc
}

type MqConn struct {
	vector      *Vector
	connI       conn.ConnI
	connVersion int32
}

func (mc *MqConn) WriteMqMsg(mqMsg *MqMsg) error {
	log.Fatal("mqMsg = %+v", proto.MarshalTextString(mqMsg))

	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if err := mc.Write(mqMsgBytes); err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) Write(bytes []byte) error {
	if err := mc.WriteSize(len(bytes)); err != nil {
		return err
	}

	err := mc.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) WriteSize(size int) error {
	var err error
	bytes, err := utility.IntToBytes(size)
	if err != nil {
		return err
	}

	err = mc.WriteBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) WriteBytes(bytes []byte) error {
	size := len(bytes)
	for size > 0 {
		wSize, err := mc.connI.Write(bytes)
		if err != nil {
			log.Fatal(err.Error())
			return err
		}

		if wSize == size {
			break
		}
		bytes = bytes[wSize:]
		size -= wSize
	}
	return nil
}

func (mc *MqConn) ReadMqMsg(timeout time.Duration) (*MqMsg, error) {
	for {
		rMqMsg, err := mc.getMqMsg()
		if err != nil {
			return nil, err
		}

		if rMqMsg != nil {
			return rMqMsg, nil
		}

		if err := mc.vector.Read(timeout); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (mc *MqConn) getMqMsgSize() (int, error) {
	sizeBytes := mc.vector.Get(4)
	if sizeBytes == nil {
		return 0, nil
	}

	size, err := utility.BytesToInt(sizeBytes)
	if err != nil {
		return 0, err
	}

	if !mc.vector.CheckSize(int(4 + size)) {
		return 0, nil
	}

	mc.vector.Skip(4)
	return int(size), nil
}

func (mc *MqConn) getMqMsg() (*MqMsg, error) {
	size, err := mc.getMqMsgSize()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}
	msgByte := mc.vector.Get(size)
	if msgByte == nil {
		return nil, nil
	}

	msg := &MqMsg{}
	if err := proto.Unmarshal(msgByte, msg); err != nil {
		return nil, log.Panic(errorCode.NewErrCode(0, err.Error()))
	}
	mc.vector.Skip(size)
	log.Fatal("msg = %+v", proto.MarshalTextString(msg))
	return msg, nil
}
