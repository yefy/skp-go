package encodes

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"

	"github.com/golang/protobuf/proto"
)

const (
	EncodeGob int32 = iota
	EncodeProto
	EncodeJson
)

func EncodeBody(encode int32, body interface{}) (string, error) {
	if encode == EncodeGob {
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)
		if err := enc.Encode(body); err != nil {
			return "", log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return buff.String(), nil

	} else if encode == EncodeProto {
		bytes, err := proto.Marshal(body.(proto.Message))
		if err != nil {
			return "", log.Panic(errorCode.NewErrCode(0, err.Error()))
		}

		return string(bytes), nil
	} else if encode == EncodeJson {
		bytes, err := json.Marshal(body)
		if err != nil {
			return "", log.Panic(errorCode.NewErrCode(0, err.Error()))
		}

		return string(bytes), nil
	}

	return "", log.Panic(errorCode.NewErrCode(0, "not encode = %+v", encode))
}

func DecodeBody(encode int32, str string, body interface{}) error {
	if encode == EncodeGob {
		var buff bytes.Buffer
		if _, err := buff.WriteString(str); err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		dec := gob.NewDecoder(&buff)
		if err := dec.Decode(body); err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return nil
	} else if encode == EncodeProto {
		err := proto.Unmarshal([]byte(str), body.(proto.Message))
		if err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return nil
	} else if encode == EncodeJson {
		err := json.Unmarshal([]byte(str), body)
		if err != nil {
			return log.Panic(errorCode.NewErrCode(0, err.Error()))
		}
		return nil
	}

	return log.Panic(errorCode.NewErrCode(0, "not encode = %+v", encode))
}
