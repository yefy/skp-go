package skpUtility

import (
	"testing"
)

func Benchmark_SkpRemoveSliceElement(b *testing.B) {
	for i := 0; i < 10000; i++ {
		intSlice := make([]int, 0)
		intSlice = append(intSlice, 1)
		intSlice = SkpRemoveSliceElement(interface{}(intSlice), 0).([]int)
		if len(intSlice) != 0 {
			b.Error("%d \n", b)
		}
	}
}

// go test -v -bench=".*"

//进行所有go文件的benchmark测试 go test -bench=".*" 或 go test . -bench=".*"
//对某个go文件进行benchmark测试 go test mysql_b_test.go -bench=".*"
//5.用性能测试生成CPU状态图（暂未测试使用）

//使用命令：
//go test -bench=".*" -cpuprofile=cpu.prof -c
//cpuprofile是表示生成的cpu profile文件
//-c是生成可执行的二进制文件，这个是生成状态图必须的，它会在本目录下生成可执行文件skpUtility.test
//然后使用go tool pprof工具
//go tool pprof skpUtility.test cpu.prof
//调用web（需要安装graphviz）来生成svg文件，生成后使用浏览器查看svg文件
