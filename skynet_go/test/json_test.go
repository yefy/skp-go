package test

import (
	"encoding/json"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq"
	"testing"

	"github.com/golang/protobuf/proto"
)

func Test_Json1(t *testing.T) {
	log.SetLevel(log.Lerr)

	sMqMsg := mq.MqMsg{}
	sMqMsg.Typ = proto.Int32(1)
	sMqMsg.Harbor = proto.Int32(1)
	sMqMsg.Topic = proto.String("topic")
	sMqMsg.Tag = proto.String("tag")
	sMqMsg.Order = proto.Uint64(1)
	sMqMsg.Class = proto.String("class")
	sMqMsg.Method = proto.String("method")
	sMqMsg.PendingSeq = proto.Uint64(1)
	sMqMsg.Encode = proto.Int32(1)
	sMqMsg.Body = proto.String("reqBody")
	log.Fatal("sMqMsg = %+v", &sMqMsg)

	b, err := json.Marshal(&sMqMsg)
	if err != nil {
		t.Error()
	}
	rMqMsg := &mq.MqMsg{}
	err = json.Unmarshal(b, rMqMsg)
	if err != nil {
		t.Error()
	}

	log.Fatal("rMqMsg = %+v", rMqMsg)
}

type JsonTest struct {
	N1    int
	N2    int
	Str1  string
	Str2  string
	StrP1 *string
	StrP2 *string
	Byte1 []byte
	Map1  map[int]string
	MapP1 *map[int]string
}

func (g *JsonTest) GetN1() int {
	return g.N1
}

func Test_Json2(t *testing.T) {
	var inGTest JsonTest
	inGTest.N1 = 1
	inGTest.N2 = 2
	inGTest.Str1 = "111"
	inGTest.Str2 = "222"
	inGTest.StrP1 = new(string)
	inGTest.StrP2 = new(string)
	*inGTest.StrP1 = "StrP1"
	*inGTest.StrP2 = "StrP2"
	inGTest.Byte1 = []byte{'1', '2', '3', '4'}
	inGTest.Map1 = map[int]string{1: "111", 2: "222"}
	inGTest.MapP1 = &map[int]string{1: "111aaa", 2: "222bbb"}

	// Note: pointer to the interface

	b, err := json.Marshal(&inGTest)
	if err != nil {
		t.Error()
	}

	var outGTest JsonTest
	err = json.Unmarshal(b, &outGTest)
	if err != nil {
		t.Error()
	}

	log.Fatal("outGTest.N1 = %+v", outGTest.N1)
	log.Fatal("outGTest.N1 = %+v", outGTest.GetN1())
	log.Fatal("outGTest.N2 = %+v", outGTest.N2)
	log.Fatal("outGTest.Str1 = %+v", outGTest.Str1)
	log.Fatal("outGTest.Str2 = %+v", outGTest.Str2)
	log.Fatal("outGTest.StrP1 = %+v", *outGTest.StrP1)
	log.Fatal("outGTest.StrP2 = %+v", *outGTest.StrP2)
	log.Fatal("outGTest.Byte1 = %+v", string(outGTest.Byte1))
	for k, v := range outGTest.Map1 {
		log.Fatal("k = %+v, v = %+v", k, v)
	}
	for k, v := range *outGTest.MapP1 {
		log.Fatal("k = %+v, v = %+v", k, v)
	}

	log.Fatal("outGTest = %+v", &outGTest)
}
