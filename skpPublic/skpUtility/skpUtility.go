package skpUtility

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	Print_Log = 0
)

func SkpRemoveSliceElement(v interface{}, index int) interface{} {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("wrong type")
	}
	slice := reflect.ValueOf(v)
	if index < 0 || index >= slice.Len() {
		panic("out of bounds")
	}
	prev := slice.Index(index)
	for i := index + 1; i < slice.Len(); i++ {
		value := slice.Index(i)
		prev.Set(value)
		prev = value
	}
	return slice.Slice(0, slice.Len()-1).Interface()
}

func Print(a ...interface{}) (n int, err error) {
	if Print_Log == 0 {
		return 0, errors.New("printf close")
	}
	return fmt.Print(a...)
}

func Printf(format string, a ...interface{}) (n int, err error) {
	if Print_Log == 0 {
		return 0, errors.New("printf close")
	}
	return fmt.Printf(format, a...)
}

func Println(a ...interface{}) (n int, err error) {
	if Print_Log == 0 {
		return 0, errors.New("printf close")
	}
	return fmt.Println(a...)
}
