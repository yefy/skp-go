package service

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
)

type ErrCode struct {
	code  int
	msg   string
	where string
}

func (e *ErrCode) Error() string {
	strCode := strconv.Itoa(e.code)
	return "ErrCode : " + e.where + " " + strCode + " " + e.msg
}

func NewErrorCode(code int, format string, a ...interface{}) error {
	pc, file, line, _ := runtime.Caller(1)
	fmt.Println(runtime.FuncForPC(pc).Name(), file, line)
	var calldepth = 1
	fmt.Println(runtime.Caller(calldepth))
	fmt.Println(string(debug.Stack()))
	errCode := &ErrCode{code, fmt.Sprintf(format, a...), ""}
	return errCode
}

type ServiceTest struct {
	Num int
}

func NewServiceTest() *ServiceTest {
	fmt.Println("NewServiceTest")
	serviceTest := &ServiceTest{}
	return serviceTest
}
func (self *ServiceTest) Test1(in int, out *int) error {
	*out = in
	if self.Num != in {
		panic(NewErrorCode(1, "Num:%+v != in:%+v", self.Num, in))
	}
	fmt.Printf("test1 in = %+v, out = %+v, Num = %+v \n", in, *out, self.Num)
	return nil
}

func (self *ServiceTest) testErr1(in int, out *int) error {
	*out = 1
	if self.Num != in {
		panic(NewErrorCode(1, "Num:%+v != in:%+v", self.Num, in))
	}
	fmt.Printf("test1 in = %+v, out = %+v, Num = %+v \n", in, *out, self.Num)
	return NewErrorCode(2, "ErrCode test")
}

func Test_Test1(t *testing.T) {
	serviceTest := NewServiceTest()
	if true {
		serviceTest.Num = 1
		in := serviceTest.Num
		out := 0
		err := serviceTest.Test1(in, &out)
		if err != nil {
			t.Error(err.Error())
		}
	}
	if true {
		serviceTest.Num = 2
		in := serviceTest.Num
		out := 0
		err := serviceTest.testErr1(in, &out)

		if err != nil {
			t.Error(err.Error())
		}
	}
	/*
		serviceTest := NewServiceTest()
		service := service.NewService(1, 1, serviceTest)
		serviceTest.Num = 1
		in := serviceTest.Num
		out := 0
		err := service.Call("test1", in, &out)
		if err != nil {
			t.Error("%+v \n", err.Error())
		}

		if in == out {
			t.Error("in:%+v != out:%+v \n", in, out)
		}
	*/
}

//go test

//测试所有的文件 go test，将对当前目录下的所有*_test.go文件进行编译并自动运行测试。
//测试某个文件使用”-file”参数。go test –file *.go 。例如：go test -file mysql_test.go，"-file"参数不是必须的，可以省略，如果你输入go test b_test.go也会得到一样的效果。
//测试某个方法 go test -run='Test_xxx'
//"-v" 参数 go test -v ... 表示无论用例是否测试通过都会显示结果，不加"-v"表示只显示未通过的用例结果
