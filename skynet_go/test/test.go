package test

import (
	"fmt"
)

type Test struct {
}

func NewTest() *Test {
	fmt.Println("NewTest")
	test := &Test{}
	return test
}

func NewTestArgv1(argv1 int) *Test {
	fmt.Println("NewTestArgv1 argv1 = ", argv1)
	test := &Test{}
	return test
}

func NewTestArgv2(argv1 int, argv2 string) *Test {
	fmt.Println("NewTestArgv1 argv1 = ", argv1, "argv2 = ", argv2)
	test := &Test{}
	return test
}

func (self *Test) Print() {
	fmt.Println("Test Print")
}

func (self *Test) Print1(argv1 int) {
	fmt.Println("Test Print argv1 = ", argv1)
}

func (self *Test) Print2(argv1 int, argv2 string) {
	fmt.Println("Test Print argv1 = ", argv1, "argv2 = ", argv2)
}

func (self *Test) RetArgv() {
	fmt.Println("Test retArgv")
}

func (self *Test) RetArgv1(argv1 int) int {
	fmt.Println("Test retArgv1 argv1 = ", argv1)
	return argv1
}

func (self *Test) RetArgv2(argv1 int, argv2 string) (int, string) {
	fmt.Println("Test retArgv2 argv1 = ", argv1, "argv2 = ", argv2)
	return argv1, argv2
}

func (self *Test) RetArgvAddr2(argv1 int, argv2 int) (int, int) {
	argv2 = 123
	fmt.Printf("Test retArgv1 argv1 = %+v, argv2 = %+v \n", argv1, argv2)
	return argv1, argv2
}

func (self *Test) RetArgvAddrX2(argv1 int, argv2 *int) (int, int) {
	*argv2 = 123
	fmt.Printf("Test retArgv1 argv1 = %+v, argv2 = %+v \n", argv1, *argv2)
	return argv1, *argv2
}
