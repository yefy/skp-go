package rpc

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"testing"
)

type ServerTest struct {
	data int
}

var errExampleFailed error = errorCode.NewErrCode(0, "ExampleFailed")

func NewServerTest() *ServerTest {
	serverTest := &ServerTest{}
	return serverTest
}
func (serverTest *ServerTest) ExampleSuccess(in int, out *int) error {
	*out = in
	if serverTest.data != in {
		return log.Panic(errorCode.NewErrCode(0, "data:%+v != in:%+v", serverTest.data, in))
	}
	//log.Debug("in = %+v, out = %+v, data = %+v \n", in, *out, serverTest.data)
	return nil
}

func (serverTest *ServerTest) ExampleFailed(in int, out *int) error {
	*out = in
	if serverTest.data != in {
		return log.Panic(errorCode.NewErrCode(0, "data:%+v != in:%+v", serverTest.data, in))
	}
	//log.Debug("in = %+v, out = %+v, data = %+v \n", in, *out, serverTest.data)
	return errExampleFailed
}

func Test_ExampleSuccess(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	in := 1
	out := 0
	serverTest.data = in
	err := serverTest.ExampleSuccess(in, &out)
	if err != nil {
		t.Error()
	}
	if in != out {
		t.Error()
	}
}

func Test_ExampleFailed(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	in := 2
	out := 0
	serverTest.data = in
	err := serverTest.ExampleFailed(in, &out)
	if in != out {
		t.Error()
	}
	if err == nil {
		t.Error()
	}
	if err != errExampleFailed {
		t.Error()
	}

}

func Test_ExampleSuccess_ExampleFailed(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	if true {
		in := 1
		out := 0
		serverTest.data = in
		err := serverTest.ExampleSuccess(in, &out)
		if err != nil {
			t.Error()
		}
		if in != out {
			t.Error()
		}
	}
	if true {
		in := 2
		out := 0
		serverTest.data = in
		err := serverTest.ExampleFailed(in, &out)
		if in != out {
			t.Error()
		}
		if err == nil {
			t.Error()
		}
		if err != errExampleFailed {
			t.Error()
		}
	}
}

func Test_ExampleSuccess_Call(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	server := NewServer(1, 1, serverTest)

	in := 1
	out := 0
	serverTest.data = in
	err := server.Call("ExampleSuccess", in, &out)
	if err != nil {
		t.Error()
	}
	if in != out {
		t.Error()
	}
	server.Stop()
}

func Test_ExampleFailed_Call(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	server := NewServer(1, 1, serverTest)

	in := 2
	out := 0
	serverTest.data = in
	err := server.Call("ExampleFailed", in, &out)
	if in != out {
		t.Error()
	}
	if err == nil {
		t.Error()
	}
	if err != errExampleFailed {
		t.Error()
	}
	server.Stop()
}

func Test_ExampleSuccess_ExampleFailed_Call(t *testing.T) {
	log.SetLevel(log.Lnone)
	serverTest := NewServerTest()
	server := NewServer(1, 1, serverTest)
	if true {
		in := 1
		out := 0
		serverTest.data = in
		err := server.Call("ExampleSuccess", in, &out)
		if err != nil {
			t.Error()
		}
		if in != out {
			t.Error()
		}
	}
	if true {
		in := 2
		out := 0
		serverTest.data = in
		err := server.Call("ExampleFailed", in, &out)
		if in != out {
			t.Error()
		}
		if err == nil {
			t.Error()
		}
		if err != errExampleFailed {
			t.Error()
		}
	}
	server.Stop()
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
