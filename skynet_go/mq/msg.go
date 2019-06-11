package mq

import (
	"skp-go/skynet_go/encodes"

	"github.com/golang/protobuf/proto"
)

func ReplyMqMsg(harbor int32, pendingSeq uint64, encode int32, body interface{}) (*MqMsg, error) {
	var str string
	str, ok := body.(string)
	if !ok {
		var err error
		str, err = encodes.EncodeBody(encode, body)
		if err != nil {
			return nil, err
		}
	}

	mqMsg := &MqMsg{}
	mqMsg.Typ = proto.Int32(TypeRespond)
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
