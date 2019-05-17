package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"skp-go/skynet_go/service"
	"skp-go/skynet_go/test"
	"strconv"
	"time"
)

type MyType struct {
	i    int
	name string
}

func (mt *MyType) SetI(i int) {
	mt.i = i
}

func (mt *MyType) SetName(name string) {
	mt.name = name
}

func (mt *MyType) String(n int, x int) (string, string) {
	return fmt.Sprintf("%p", mt) + "--name:" + mt.name + " i:" + strconv.Itoa(mt.i) + "n :" + strconv.Itoa(n), "678"
}

func args(ag ...interface{}) []interface{} {
	return ag
}

func add(n int, n2 int) {

}

func tsfaf(ag ...interface{}) {
	for key, value := range ag {
		fmt.Println(key, value)
	}
}

var logg *log.Logger

func someHandler() {
	ctx, cancel := context.WithCancel(context.Background())
	go doStuff1_1(ctx)
	go doStuff1_2(ctx)
	go doStuff1_3(ctx)

	//10秒后取消doStuff
	time.Sleep(3 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

}

func doStuff1_1_1(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			logg.Printf("done1_1")
			return
		default:
			logg.Printf("work1_1")
		}
	}
}

func doStuff1_1_2(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			logg.Printf("done1_2")
			return
		default:
			logg.Printf("work1_2")
		}
	}
}

func doStuff1_1_3(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			logg.Printf("done1_3")
			return
		default:
			logg.Printf("work1_3")
		}
	}
}

//每1秒work一下，同时会判断ctx是否被取消了，如果是就退出
func doStuff1_1(ctx context.Context) {

	ctx1, _ := context.WithCancel(ctx)
	go doStuff1_1_1(ctx1)
	go doStuff1_1_2(ctx1)
	go doStuff1_1_3(ctx1)

	//var timen int = 1
	for {
		time.Sleep(1 * time.Second)
		//		timen = timen + 1
		//		if timen > 3 {
		//			cancel1()
		//		}
		select {
		case <-ctx.Done():
			logg.Printf("done1")
			return
		default:
			logg.Printf("work1")
		}
	}
}

func doStuff1_2(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			logg.Printf("done2")
			return
		default:
			logg.Printf("work2")
		}
	}
}

func doStuff1_3(ctx context.Context) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			logg.Printf("done3")
			return
		default:
			logg.Printf("work3")
		}
	}
}

func main1() {
	logg = log.New(os.Stdout, "", log.Ltime)
	someHandler()
	logg.Printf("down")
}

func fTest2(data interface{}) {
	fmt.Printf("data = %+v \n", data)
}

func fTest() {
	f1 := func() {
		fmt.Println("ffff")
	}
	fTest2(f1)
	f2 := func() {
		fmt.Println("ffff")
	}
	fTest2(f2)

	fTest2(func() {
		fmt.Println("ffff")
	})

	fTest2(func() {
		fmt.Println("ffff")
	})
}

func counter(start int) (func() int, func()) {
	// if the value gets mutated, the same is reflected in closure
	ctr := func() int {
		return start
	}

	incr := func() {
		start++
	}

	// both ctr and incr have same reference to start
	// closures are created, but are not called
	return ctr, incr
}

func testNewService(service *service.Service, num int) {
	addrNum := 1

	service.Call("RetArgvAddrX2", num, &addrNum)
	fmt.Printf("ffff addrNum = %+v \n", addrNum)
}

func main() {
	n := 0
	t := test.NewTest()
	testService := service.NewService(1, 1, t)
	err := testService.Call("ArgStru", 1, &n)
	if err == nil {
		fmt.Printf("n = %+v \n", n)
	} else {
		fmt.Printf("err = %+v \n", err.Error())
	}

	if false {
		main1()
	}

	if false {
		for i := 0; i < 2; i++ {
			fTest()
		}
		n := 1
		x := 2
		_ = n
		_ = x
		for i := 0; i < 2; i++ {
			fTest()
		}
	}

	if false {
		// ctr, incr and ctr1, incr1 are different
		ctr, incr := counter(100)
		ctr1, incr1 := counter(100)
		fmt.Println("counter - ", ctr())
		fmt.Println("counter1 - ", ctr1())
		// incr by 1
		incr()
		fmt.Println("counter - ", ctr())
		fmt.Println("counter1- ", ctr1())
		// incr1 by 2
		incr1()
		incr1()
		fmt.Println("counter - ", ctr())
		fmt.Println("counter1- ", ctr1())
	}

	if false {
		num := 1
		testService := service.NewService(1, 1, test.NewTest)
		testNewService(testService, num)
		num++
		testNewService(testService, num)
		num++
		testNewService(testService, num)
		num++
		testNewService(testService, num)
		num++
		testNewService(testService, num)
	}

	if false {
		num := 0
		for i := 0; i < 2; i++ {
			num++
			fmt.Printf("main index = %+v \n", i)
			testService := service.NewService(1, 1, test.NewTest)
			testService.Send("Print")
			testService.Send("Print1", 1)
			testService.Send("Print2", 1, "2")

			testService.Call("RetArgv", 1, 1)

			testService.Call("RetArgv1", 1, 1)

			testService.Call("RetArgv2", 1, "2")

			//	testService1 := service.NewService(test.NewTestArgv1, 1)
			//	testService1.Send("Print")
			//	testService1.Send("Print1", 1)
			//	testService1.Send("Print2", 1, "2")
			//	testService2 := service.NewService(test.NewTestArgv2, 1, "2")
			//	testService2.Send("Print")
			//	testService2.Send("Print1", 1)
			//	testService2.Send("Print2", 1, "2")
			fmt.Printf("end index = %+v \n", i)
			testService.Stop()
		}
	}

	if false {

		if true {
			vType := reflect.TypeOf(add)
			numIn := vType.NumIn()
			fmt.Println("add numIn = ", numIn)
			//返回func类型的参数个数，如果不是函数，将会panic
			addIn := make([]reflect.Type, numIn)
			for i := 0; i < numIn; i++ {
				addIn[i] = vType.In(i)
				//返回func类型的第i个参数的类型，如非函数或者i不在[0, NumIn())内将会panic
				fmt.Println("add ag = ", addIn[i])
			}
		}

		myType := &MyType{22, "wowzai"}
		//fmt.Println(myType)     //就是检查一下myType对象内容
		//println("---------------")
		mtV := reflect.ValueOf(&myType).Elem()
		paramsxx := make([]reflect.Value, 1)
		paramsxx[0] = reflect.ValueOf(79)

		ag := args(489, 789)
		fmt.Println("ag len = ", len(ag))
		paramsxxag := make([]reflect.Value, 2)
		paramsxxag[0] = reflect.ValueOf(ag[0])
		paramsxxag[1] = reflect.ValueOf(ag[1])

		agf := mtV.MethodByName("String").Type()
		aafa := agf.NumIn()
		fmt.Println("aafa1111111 = ", aafa)
		//strvType := reflect.TypeOf(agf)
		//strnumIn := strvType.NumIn()
		//fmt.Println("String strnumIn = ", strnumIn)

		xxc := mtV.MethodByName("String").Call(paramsxxag)
		fmt.Println("Before:", xxc[0], xxc[1])
		//fmt.Println("Before:", mtV.MethodByName("String").Call(nil)[0])

		if true {
			f := func(a string, b string) {
				fmt.Println("ffff = ", a, b)
			}

			// 将函数包装为反射值对象
			funcValue := reflect.ValueOf(f)
			// 构造函数参数, 传入两个整型值
			paramList := []reflect.Value{reflect.ValueOf(xxc[0].Interface()), reflect.ValueOf(xxc[1].Interface())}
			// 反射调用函数
			funcValue.Call(paramList)

		}

		params := make([]reflect.Value, 1)
		params[0] = reflect.ValueOf(18)
		mtV.MethodByName("SetI").Call(params)
		params[0] = reflect.ValueOf("reflection test")
		mtV.MethodByName("SetName").Call(params)
		//fmt.Println("After:", mtV.MethodByName("String").Call(paramsxx)[0])
	}
}
