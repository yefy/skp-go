package rpc

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"testing"
)

var _ = errorCode.NewErrCode

type ServerTest struct {
	n1    int
	n2    int
	str1  string
	str2  string
	pn1   *int
	pn2   *int
	pstr1 *string
	pstr2 *string
}

func (s *ServerTest) compare(c *ServerTest) bool {
	if s.n1 != c.n1 ||
		s.n2 != c.n2 ||
		s.str1 != c.str1 ||
		s.str2 != c.str2 ||
		*s.pn1 != *c.pn1 ||
		*s.pn2 != *c.pn2 ||
		*s.pstr1 != *c.pstr1 ||
		*s.pstr2 != *c.pstr2 {
		return false

	}
	return true
}

func (s *ServerTest) clone() *ServerTest {
	server := &ServerTest{}
	server.n1 = s.n1
	server.n2 = s.n2
	server.str1 = s.str1
	server.str2 = s.str2

	server.pn1 = new(int)
	server.pn2 = new(int)
	server.pstr1 = new(string)
	server.pstr2 = new(string)

	*server.pn1 = *s.pn1
	*server.pn2 = *s.pn2
	*server.pstr1 = *s.pstr1
	*server.pstr2 = *s.pstr2
	return server
}

func (s *ServerTest) data() *ServerTest {
	server := &ServerTest{}
	server.n1 = 1
	server.n2 = 2
	server.str1 = "str1"
	server.str2 = "str2"

	server.pn1 = new(int)
	server.pn2 = new(int)
	server.pstr1 = new(string)
	server.pstr2 = new(string)

	*server.pn1 = 11
	*server.pn2 = 22
	*server.pstr1 = "pstr1"
	*server.pstr2 = "pstr2"
	return server
}

func NewServerTest() *ServerTest {
	serverTest := &ServerTest{}
	return serverTest
}

func (serverTest *ServerTest) ExampleTest(in *ServerTest, out *ServerTest) (*ServerTest, error) {
	log.All("in = %+v", in)
	*out = *in.clone()
	//return out, errorCode.NewErrCode(0, "ExampleTest")
	return out, nil
}

func (serverTest *ServerTest) ExampleTestError(in *ServerTest, out *ServerTest) error {
	log.All("in = %+v", in)
	*out = *in.clone()
	//return out
	return errorCode.NewErrCode(0, "ExampleTest")
}

func (serverTest *ServerTest) ExamplePerf(in *int, out *int) (*int, *int) {
	*out = *in
	return in, out
}

func Test_ExampleTest_Send(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := serverTest.data()
	out := ServerTest{}
	err := server.Send("ExampleTest", in, &out)
	if err != nil {
		t.Error()
	}

	server.Stop(true)
}

func Test_ExampleTest_SendReq(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := serverTest.data()
	out := ServerTest{}

	err := server.SendReq("ExampleTest", in, &out, func(resServer *ServerTest, resErr error) {
		if in.compare(&out) == false {
			t.Error()
		}

		if in.compare(resServer) == false {
			t.Error()
		}

		if resErr != nil {
			t.Error()
		}

	})
	if err != nil {
		t.Error()
	}

	server.Stop(true)
}

func Test_ExampleTest_Call(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := serverTest.data()
	out := ServerTest{}
	err := server.Call("ExampleTest", in, &out)
	if err != nil {
		t.Error()
	}

	if in.compare(&out) == false {
		t.Error()
	}

	server.Stop(true)
}

func Test_ExampleTestError_Call(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := serverTest.data()
	out := ServerTest{}
	err := server.Call("ExampleTestError", in, &out)
	if err == nil {
		t.Error()
	}

	if in.compare(&out) == false {
		t.Error()
	}

	server.Stop(true)
}

func Test_ExampleTest_CallReq(t *testing.T) {
	log.SetLevel(log.Lerr)
	serverTest := NewServerTest()
	server := NewServer(serverTest)

	in := serverTest.data()
	out := ServerTest{}

	err := server.CallReq("ExampleTest", in, &out, func(resServer *ServerTest, resErr error) {
		if in.compare(&out) == false {
			t.Error()
		}

		if in.compare(resServer) == false {
			t.Error()
		}

		if resErr != nil {
			t.Error()
		}

	})
	if err != nil {
		t.Error()
	}

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
