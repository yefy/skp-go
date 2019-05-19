package errorCode

import (
	log "skp-go/skynet_go/logger"
	"testing"
)

func Test_NewErrCode(t *testing.T) {
	errCode := NewErrCode(0, "Test_NewErrCode")
	err := errCode.Error()
	log.Fatal(err)
}

func Test_NewErrCodeWhere(t *testing.T) {
	errCode := NewErrCodeWhere("skynet_go", 1, "Test_NewErrCodeWhere")
	err := errCode.Error()
	log.Fatal(err)
}

//go test

//测试所有的文件 go test，将对当前目录下的所有*_test.go文件进行编译并自动运行测试。
//测试某个文件使用”-file”参数。go test –file *.go 。例如：go test -file mysql_test.go，"-file"参数不是必须的，可以省略，如果你输入go test b_test.go也会得到一样的效果。
//测试某个方法 go test -run='Test_xxx'
//"-v" 参数 go test -v ... 表示无论用例是否测试通过都会显示结果，不加"-v"表示只显示未通过的用例结果
