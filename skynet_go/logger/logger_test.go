package logger

import (
	_ "errors"
	"testing"
)

func Test_Logger(t *testing.T) {
	SetLevel(Lall)
	All("Test_Logger")
	Debug("Test_Logger")
	Trace("Test_Logger")
	Err("Test_Logger")
	Fatal("Test_Logger")
	Print(Lfatal, 0, "Test_Logger")
	//Panic(errors.New("Test_Logger"))
}

func Test_NewGlobalLogger(t *testing.T) {
	NewGlobalLogger("./testGlobal.log", "testGlobal", Ldate|Lmicroseconds|Lshortfile, Lall, Lscreen|Lfile)
	All("Test_NewGlobalLogger")
	Debug("Test_NewGlobalLogger")
	Trace("Test_NewGlobalLogger")
	Err("Test_NewGlobalLogger")
	Fatal("Test_NewGlobalLogger")
	Print(Lfatal, 0, "Test_NewGlobalLogger")
	//Panic(errors.New("Test_NewGlobalLogger"))
}

func Test_New(t *testing.T) {
	log := New("./test.log", "test", Ldate|Lmicroseconds|Lshortfile, Lall, Lscreen|Lfile)
	log.All("Test_New")
	log.Debug("Test_New")
	log.Trace("Test_New")
	log.Err("Test_New")
	log.Fatal("Test_New")
	log.Print(Lfatal, 0, "Test_New")
	//log.Panic(errors.New("Test_New"))
}

//go test

//测试所有的文件 go test，将对当前目录下的所有*_test.go文件进行编译并自动运行测试。
//测试某个文件使用”-file”参数。go test –file *.go 。例如：go test -file mysql_test.go，"-file"参数不是必须的，可以省略，如果你输入go test b_test.go也会得到一样的效果。
//测试某个方法 go test -run='Test_xxx'
//"-v" 参数 go test -v ... 表示无论用例是否测试通过都会显示结果，不加"-v"表示只显示未通过的用例结果
