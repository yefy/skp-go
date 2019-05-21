// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log implements a simple logging package. It defines a type, Logger,
// with methods for formatting output. It also has a predefined 'standard'
// Logger accessible through helper functions Print[f|ln], Fatal[f|ln], and
// Panic[f|ln], which are easier to use than creating a Logger manually.
// That logger writes to standard error and prints the date and time
// of each logged message.
// Every log message is output on a separate line: if the message being
// printed does not end in a newline, the logger will add one.
// The Fatal functions call os.Exit(1) after writing the log message.
// The Panic functions call panic after writing the log message.
package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// These flags define which text to prefix to each log entry generated by the Logger.
const (
	// Bits or'ed together to control what's printed.
	// There is no control over the order they appear (the order listed
	// here) or the format they present (as described in the comments).
	// The prefix is followed by a colon only when Llongfile or Lshortfile
	// is specified.
	// For example, flags Ldate | Ltime (or LstdFlags) produce,
	//	2009/01/23 01:23:23 message
	// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
	//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
	Ldate         = 1 << iota                          // the date in the local time zone: 2009/01/23
	Ltime                                              // the time in the local time zone: 01:23:23
	Lmicroseconds                                      // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                                          // full file name and line number: /a/b/c/d.go:23
	Lshortfile                                         // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                                               // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Lmicroseconds | Lshortfile // initial values for the standard logger
)

const (
	Lall = iota
	Ldebug
	Ltrace
	Lerr
	Lfatal
	Lnone
	LlevelFlags = Lerr
)

var LlevelName = []string{"  all", "debug", "trace", "  err", "fatal"}

const (
	Lscreen = 1 << iota
	Lfile
	LmodelFlags = Lscreen
)

// A Logger represents an active logging object that generates lines of
// output to an io.Writer. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write
	level  int
	model  int
	file   *os.File
}

// New creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.

func New(fullPath string, prefix string, flag int, level int, model int) *Logger {
	var file *os.File
	if model&Lfile != 0 {
		var err error
		file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("打开日志文件失败(开启屏幕打印)：", err)
			model = Lscreen
		}
	}

	var out io.Writer
	if model&Lscreen != 0 && model&Lfile != 0 {
		out = io.MultiWriter(os.Stdout, file)
	} else if model&Lscreen != 0 {
		out = os.Stdout
	} else if model&Lfile != 0 {
		out = file
	}

	return &Logger{
		prefix: prefix,
		flag:   flag,
		out:    out,
		level:  level,
		model:  model,
		file:   file,
	}
}

//var std = New("./global.log", "", LstdFlags, LlevelFlags, LmodelFlags)
var std = New("", "", LstdFlags, LlevelFlags, LmodelFlags)

func NewGlobalLogger(fullPath string, prefix string, flag int, level int, model int) {
	std = New(fullPath, prefix, flag, level, model)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *Logger) GetGID() string {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b[:])
	// n, _ := strconv.ParseUint(string(b), 10, 64)
	// return n
}

// FormatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (l *Logger) FormatHeader(t time.Time, funcName string, file string, line int, levelName string) {
	l.buf = l.buf[:0]
	buf := &l.buf
	*buf = append(*buf, l.prefix...)
	if len(l.prefix) > 0 {
		*buf = append(*buf, " | "...)
	}
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			//*buf = append(*buf, ' ')
		}
	}
	*buf = append(*buf, " | "...)
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}

		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, " | "...)
		*buf = append(*buf, funcName...)
	}

	id := l.GetGID()
	if len(id) > 0 {
		*buf = append(*buf, " | "...)
		*buf = append(*buf, id...)
	}

	if len(levelName) > 0 {
		*buf = append(*buf, " | "...)
		*buf = append(*buf, levelName...)
	}

}

func FormatHeader(t time.Time, funcName string, file string, line int, levelName string) string {
	std.FormatHeader(t, funcName, file, line, levelName)
	return string(std.buf[:])
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *Logger) Output(level int, calldepth int, format string, v ...interface{}) error {
	if level < l.level {
		return nil
	}

	s := fmt.Sprintf(format+"\n", v...)

	levelName := LlevelName[level]

	now := time.Now() // get this early.
	var funcName string
	var pc uintptr
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		pc, file, line, ok = runtime.Caller(calldepth)
		funcName = runtime.FuncForPC(pc).Name()
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}

	l.FormatHeader(now, funcName, file, line, levelName)

	l.buf = append(l.buf, " | "...)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(level int, calldepth int, s string) error {
	if level < std.level {
		return nil
	}
	return std.Output(level, calldepth+1, s) // +1 for this frame.
}

func (l *Logger) All(format string, v ...interface{}) {
	if Lall < l.level {
		return
	}
	l.Output(Lall, 2, format, v...)
}

func All(format string, v ...interface{}) {
	if Lall < std.level {
		return
	}
	std.Output(Lall, 2, format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if Ldebug < l.level {
		return
	}
	l.Output(Ldebug, 2, format, v...)
}

func Debug(format string, v ...interface{}) {
	if Ldebug < std.level {
		return
	}
	std.Output(Ldebug, 2, format, v...)
}

func (l *Logger) Trace(format string, v ...interface{}) {
	if Ltrace < l.level {
		return
	}
	l.Output(Ltrace, 2, format, v...)
}

func Trace(format string, v ...interface{}) {
	if Ltrace < std.level {
		return
	}
	std.Output(Ltrace, 2, format, v...)
}

func (l *Logger) Err(format string, v ...interface{}) {
	if Lerr < l.level {
		return
	}
	l.Output(Lerr, 2, format, v...)
}

func Err(format string, v ...interface{}) {
	if Lerr < std.level {
		return
	}
	std.Output(Lerr, 2, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if Lfatal < l.level {
		return
	}
	l.Output(Lfatal, 2, format, v...)
}

func Fatal(format string, v ...interface{}) {
	if Lfatal < std.level {
		return
	}
	std.Output(Lfatal, 2, format, v...)
}

func (l *Logger) Print(level int, calldepth int, format string, v ...interface{}) {
	if level < l.level {
		return
	}
	calldepth = calldepth + 2
	l.Output(level, calldepth, format, v...)
}

func Print(level int, calldepth int, format string, v ...interface{}) {
	if level < std.level {
		return
	}
	calldepth = calldepth + 2
	std.Output(level, calldepth, format, v...)
}

func (l *Logger) Panic(err error) error {
	if Lerr < l.level {
		return nil
	}
	l.Output(Lerr, 2, err.Error())
	panic(err)
	return err
}

func Panic(err error) error {
	if Lerr < std.level {
		panic(err)
		return err
	}
	std.Output(Lerr, 2, err.Error())
	panic(err)
	return err
}

func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func SetLevel(level int) {
	std.SetLevel(level)
}

func (l *Logger) SetModel(model int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.model = model
}

func SetModel(model int) {
	std.SetModel(model)
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// Flags returns the output flags for the logger.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

func Flags() int {
	return std.Flags()
}

// SetFlags sets the output flags for the logger.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

func SetFlags(flag int) {
	std.SetFlags(flag)
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

func Prefix() string {
	return std.Prefix()
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}
