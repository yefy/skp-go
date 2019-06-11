package rpcEncode

import (
	"bytes"
	"encoding/gob"
	log "skp-go/skynet_go/logger"
	"skp-go/skynet_go/mq/rpcProto"
	"testing"

	protobuf "github.com/golang/protobuf/proto"
)

type ServerTest struct {
	ServerBase
}

func NewServerTest() *ServerTest {
	serverTest := &ServerTest{}
	return serverTest
}

func (serverTest *ServerTest) ExampleTest(in *rpcProto.Person, out *rpcProto.Person) error {
	log.Fatal("in = %+v", in)
	out.Name = protobuf.String(in.GetName())
	out.Age = protobuf.Int32(in.GetAge())
	out.Email = protobuf.String(in.GetEmail())

	return nil
}

func Test_ExampleTest_Send(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := &rpcProto.Person{
		Name:  protobuf.String("111"),
		Age:   protobuf.Int32(222),
		Email: protobuf.String("333"),
	}

	var bufIn bytes.Buffer
	enc := gob.NewEncoder(&bufIn)
	enc.Encode(in)
	inStr := bufIn.String()

	err := server.Send("ExampleTest", inStr)
	if err != nil {
		t.Error()
	}
	server.Stop(true)
}

func Test_ExampleTest_SendReq(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := &rpcProto.Person{
		Name:  protobuf.String("111"),
		Age:   protobuf.Int32(222),
		Email: protobuf.String("333"),
	}

	var bufIn bytes.Buffer
	enc := gob.NewEncoder(&bufIn)
	enc.Encode(in)
	inStr := bufIn.String()

	server.SendReq("ExampleTest", inStr, func(outStr string, err error) {
		if err != nil {
			t.Error()
		}
		log.Fatal("outStr = %s", outStr)
	})
	server.Stop(true)
}

func Test_ExampleTest_Call(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := &rpcProto.Person{
		Name:  protobuf.String("111"),
		Age:   protobuf.Int32(222),
		Email: protobuf.String("333"),
	}

	var bufIn bytes.Buffer
	enc := gob.NewEncoder(&bufIn)
	enc.Encode(in)
	inStr := bufIn.String()
	log.Fatal("inStr = %s", inStr)
	outStr, err := server.Call("ExampleTest", inStr)
	if err != nil {
		t.Error()
	}

	log.Fatal("outStr = %s", outStr)
	server.Stop(true)
}

//go test

//测试所有的文件 go test，将对当前目录下的所有*_test.go文件进行编译并自动运行测试。
//测试某个文件使用”-file”参数。go test –file *.go 。例如：go test -file mysql_test.go，"-file"参数不是必须的，可以省略，如果你输入go test b_test.go也会得到一样的效果。
//测试某个方法 go test -run='Test_xxx'
//"-v" 参数 go test -v ... 表示无论用例是否测试通过都会显示结果，不加"-v"表示只显示未通过的用例结果

// 使用如下命令行开启基准测试：
// $ go test -v -bench=. benchmark_test.go
// goos: linux
// goarch: amd64
// Benchmark_Add-4           20000000         0.33 ns/op
// PASS
// ok          command-line-arguments        0.700s
// 代码说明如下：
// 第 1 行的-bench=.表示运行 benchmark_test.go 文件里的所有基准测试，和单元测试中的-run类似。
// 第 4 行中显示基准测试名称，2000000000 表示测试的次数，也就是 testing.B 结构中提供给程序使用的 N。“0.33 ns/op”表示每一个操作耗费多少时间（纳秒）。

// 注意：Windows 下使用 go test 命令行时，-bench=.应写为-bench="."
