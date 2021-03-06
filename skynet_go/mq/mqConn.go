package mq

import (
	"reflect"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/utility"
	"time"

	"github.com/golang/protobuf/proto"
)

func NewMqConn() *MqConn {
	mc := &MqConn{}
	return mc
}

type MqConn struct {
	vector *Vector
	connI  ConnI
	isOk   bool
}

func (mc *MqConn) IsOk() bool {
	return mc.isOk
}

func (mc *MqConn) SetConn(connI ConnI) bool {
	if connI == nil {
		return false
	}

	defer func() {
		mc.isOk = true
	}()

	if mc.connI == nil {
		mc.connI = connI
		mc.vector = NewVector()
		mc.vector.SetConn(mc.connI)
		return true
	} else {
		if reflect.ValueOf(mc.connI).Pointer() != reflect.ValueOf(connI).Pointer() {
			mc.connI = connI
			mc.vector = NewVector()
			mc.vector.SetConn(mc.connI)
			return true
		} else {
			//mc.vector.SetConn(mc.connI)
			return false
		}
	}
	return true
}

func (mc *MqConn) GetConn() ConnI {
	return mc.connI
}

func (mc *MqConn) WriteMqMsg(mqMsg *MqMsg) error {
	log.All("mqMsg = %+v", proto.MarshalTextString(mqMsg))

	mqMsgBytes, err := proto.Marshal(mqMsg)
	if err != nil {
		log.Err("proto.Marshal(mqMsg) err = %+v", mqMsg)
		return nil
		//return log.Panic(errorCode.NewErrCode(0, err.Error()))
	}

	if err := mc.Write(mqMsgBytes); err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) Write(bytes []byte) (err error) {
	defer func() {
		if err != nil {
			mc.isOk = false
		}
	}()

	if err = mc.writeSize(len(bytes)); err != nil {
		return err
	}

	err = mc.writeBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) writeSize(size int) error {
	var err error
	bytes, err := utility.IntToBytes(size)
	if err != nil {
		return err
	}

	err = mc.writeBytes(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (mc *MqConn) writeBytes(bytes []byte) error {
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

func (mc *MqConn) ReadMqMsg(timeout time.Duration) (rMqMsg *MqMsg, err error) {
	defer func() {
		if err != nil {
			mc.isOk = false
		}
	}()

	for {
		rMqMsg, err = mc.getMqMsg()
		if err != nil {
			return nil, err
		}

		if rMqMsg != nil {
			return rMqMsg, nil
		}

		if err = mc.vector.Read(timeout); err != nil {
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
	log.All("msg = %+v", proto.MarshalTextString(msg))
	return msg, nil
}
