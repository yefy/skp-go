package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"reflect"
	log "skp-go/skynet_go/logger"
	"sync"
	"unsafe"
)

var _ = fmt.Errorf
var _ = reflect.Bool

type GobTest struct {
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

func (g *GobTest) GetN1() int {
	return g.N1
}

func StringTest() {
	var s string = "123"
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	log.Fatal("s = %+v, sh = %+v, Data = %p", s, sh, sh.Data)
	var s2 string = s + "1"
	sh2 := (*reflect.StringHeader)(unsafe.Pointer(&s2))
	log.Fatal("s2 = %+v, sh2 = %+v, Data = %p", s2, sh2, sh2.Data)
	//bh := reflect.SliceHeader{sh.Data, sh.Len, 0}
}

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

func main() {
	bbb := IntToBytes(11111111111111111)
	fmt.Println("len(bbb)", len(bbb))
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Lall, log.Lscreen|log.Lfile)
	log.All("main start")

	StringTest()

	var smap sync.Map
	// smap.Store(1, 11)
	// v1, isok1 := smap.Load(1)
	// log.Fatal("v1 = %+v, isok1 = %+v", v1, isok1)
	xxx, xxxok := smap.LoadOrStore(1, 22)
	log.Fatal("xxx = %+v, xxxok = %+v", xxx, xxxok)
	v2, isok2 := smap.Load(1)
	log.Fatal("v2 = %+v, isok2 = %+v", v2, isok2)
	smap.Store(2, 22)
	smap.Range(func(key, value interface{}) bool {
		log.Fatal("key = %+v, value = %+v", key, value)
		return true
	})

	network := new(bytes.Buffer)
	enc := gob.NewEncoder(network)

	var inGTest GobTest
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
	err := enc.Encode(&inGTest)
	if err != nil {
		panic(err)
	}

	var outGTest GobTest

	dec := gob.NewDecoder(network)

	err = dec.Decode(&outGTest)
	if err != nil {
		panic(err)
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

	log.Fatal("outGTest = %+v", outGTest)
}
