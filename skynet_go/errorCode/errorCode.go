package errorCode

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"skp-go/skynet_go/logger"
	"time"
)

const (
	Unknown int32 = iota
	TimeOut
)

var Line string = "\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
var isStack bool = false

func SetStack(b bool) {
	isStack = b
}

func GetCode(err error) int32 {
	e, ok := err.(*ErrCode)
	if !ok {
		return -1
	} else {
		return e.Code()
	}
}

type ErrCode struct {
	server string
	code   int32
	msg    string
	where  string
	stack  string
}

func (e *ErrCode) Code() int32 {
	return e.code
}

func (e *ErrCode) Error() string {
	err := fmt.Sprintf("server:(%s) errCode:(%d) where:(%s) msg:(%s)", e.server, e.code, e.where, e.msg)
	if len(e.stack) > 0 {
		err = fmt.Sprintf("%s\n*****goroutine stack start*****\n%s*****goroutinestack stack end*****%s", err, e.stack, Line)
	}
	return err
}

func getErrCode(server string, code int32, format string, a ...interface{}) error {
	pc, file, line, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()
	now := time.Now()
	levelName := ""
	where := logger.FormatHeader(now, funcName, file, line, levelName)
	var stack string
	if isStack {
		stack = string(debug.Stack())
	}
	errCode := &ErrCode{server, code, fmt.Sprintf(format, a...), where, stack}
	return errCode
}

func NewErrCode(code int32, format string, a ...interface{}) error {
	return getErrCode("", code, format, a...)
}

func NewErrCodeWhere(server string, code int32, format string, a ...interface{}) error {
	return getErrCode(server, code, format, a...)
}
