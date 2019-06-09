package mq

import (
	"bytes"
	"encoding/gob"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"

	"github.com/golang/protobuf/proto"
)

func EncodeBody(encode int32, body interface{}) (string, error) {
	if encode == encodeGob {
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)
		if err := enc.Encode(body); err != nil {
			return "", log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return buff.String(), nil

	} else if encode == encodeProto {
	} else if encode == encodeJson {
		return "", log.Panic(errorCode.NewErrCode(0, "not encodeJson"))
	}

	return "", log.Panic(errorCode.NewErrCode(0, "not encode = %+v", encode))
}

func DecodeBody(encode int32, str string, body interface{}) error {
	if encode == encodeGob {
		var buff bytes.Buffer
		if _, err := buff.WriteString(str); err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		dec := gob.NewDecoder(&buff)
		if err := dec.Decode(body); err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return nil
	} else if encode == encodeProto {
	} else if encode == encodeJson {
		return log.Panic(errorCode.NewErrCode(0, "not encodeJson"))
	}

	return log.Panic(errorCode.NewErrCode(0, "not encode = %+v", encode))
}

func ReplyMqMsg(harbor int32, pendingSeq uint64, encode int32, body interface{}) (*MqMsg, error) {
	str, err := EncodeBody(encode, body)
	if err != nil {
		return nil, err
	}

	mqMsg := &MqMsg{}
	mqMsg.Typ = proto.Int32(typRespond)
	mqMsg.Harbor = proto.Int32(harbor)
	mqMsg.PendingSeq = proto.Uint64(pendingSeq)
	mqMsg.Encode = proto.Int32(encode)
	mqMsg.Body = proto.String(str)
	mqMsg.Topic = proto.String("")
	mqMsg.Tag = proto.String("")
	mqMsg.Order = proto.Uint64(0)
	mqMsg.Class = proto.String("")
	mqMsg.Method = proto.String("")

	return mqMsg, nil
}
