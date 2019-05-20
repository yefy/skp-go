package rpc

import (
	"skp-go/skynet_go/errorCode"
	log "skp-go/skynet_go/logger"
	"testing"
)

//go test -test.bench=. server_benchmark_test.go server.go
//go test -run=xxx -bench=. -benchtime="3s" -cpuprofile profile_cpu.out
//go test -run=xxx -bench=Benchmark_Test_rpc_Send$ server_benchmark_test.go server.go -benchtime="3s" -cpuprofile profile_cpu.out
//go tool pprof app.test profile_cpu.out
//#top10
//go tool pprof -svg profile_cpu.out > profile_cpu.svg
//go tool pprof -pdf profile_cpu.out > profile_cpu.pdf

//go test -run=xxx -bench=Benchmark_Test_rpc_Send$ server_benchmark_test.go server.go -benchtime="3s" -cpuprofile profile_cpu.out
//go tool pprof -pdf profile_cpu.out > profile_cpu.pdf
//go test -run=xxx -bench=Benchmark_Test_rpc_Send$ server_benchmark_test.go server.go -benchtime="3s" -memprofile  profile_mem.out
//go tool pprof -pdf profile_mem.out > profile_mem.pdf

type ServiceTest_b struct {
	Num int
}

var errTestb error = errorCode.NewErrCode(0, "TestErr")

func NewServiceTest_b() *ServiceTest_b {
	serviceTest := &ServiceTest_b{}
	return serviceTest
}
func (serviceTest *ServiceTest_b) Test(in int, out *int) error {
	*out = in
	if serviceTest.Num != in {
		return log.Panic(errorCode.NewErrCode(0, "Test Num:%+v != in:%+v", serviceTest.Num, in))
	}
	log.Debug("Test in = %+v, out = %+v, Num = %+v \n", in, *out, serviceTest.Num)
	return nil
}

func (serviceTest *ServiceTest_b) TestErr(in int, out *int) error {
	*out = in
	if serviceTest.Num != in {
		return log.Panic(errorCode.NewErrCode(0, "TestErr Num:%+v != in:%+v", serviceTest.Num, in))
	}
	log.Debug("TestErr in = %+v, out = %+v, Num = %+v \n", in, *out, serviceTest.Num)
	return errTestb
}

func Benchmark_Test(b *testing.B) {
	log.SetLevel(log.Lnone)
	serviceTest := NewServiceTest_b()
	for i := 0; i < b.N; i++ {
		in := 1
		out := 0
		serviceTest.Num = in
		err := serviceTest.Test(in, &out)
		if err != nil {
			b.Error()
		}
		if in != out {
			b.Error()
		}
	}
}

func Benchmark_Test_rpc_call(b *testing.B) {
	b.ReportAllocs()
	log.SetLevel(log.Lnone)
	serviceTest := NewServiceTest_b()
	server := NewServer(1, 1, serviceTest)
	for i := 0; i < b.N; i++ {
		in := 1
		out := 0
		serviceTest.Num = in
		err := server.Call("Test", in, &out)
		if err != nil {
			b.Error()
		}
		if in != out {
			b.Error()
		}
	}
	server.Stop()
}

func Benchmark_Test_rpc_Send(b *testing.B) {
	b.ReportAllocs()
	log.SetLevel(log.Lnone)
	serviceTest := NewServiceTest_b()
	server := NewServer(1, 1000, serviceTest)
	for i := 0; i < b.N; i++ {
		in := 1
		out := 0
		serviceTest.Num = in
		err := server.Send("Test", in, &out)
		if err != nil {
			b.Error()
		}
	}
	server.Stop()
}

// goos: windows
// goarch: amd64
// Benchmark_Test-4                20000000                59.1 ns/op
// Benchmark_Test_rpc_call-4        1000000              2250 ns/op
// Benchmark_Test_rpc_Send-4        2000000               994 ns/op
// PASS
// ok      command-line-arguments  6.673s

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
