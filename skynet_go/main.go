package main

import (
	"bytes"
	"encoding/gob"
	log "skp-go/skynet_go/logger"
)

type GobTest struct {
	N1    int
	N2    int
	Str1  string
	Str2  string
	StrP1 *string
	StrP2 *string
	Byte1 []byte
	Map1  map[int]string
}

func (g *GobTest) GetN1() int {
	return g.N1
}

func main() {
	log.NewGlobalLogger("./global.log", "", log.LstdFlags, log.Lall, log.Lscreen|log.Lfile)
	log.All("main start")

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

	b := []byte{'1', '2', '3', '4'}
	log.Fatal("b = %+v", string(b))
}
