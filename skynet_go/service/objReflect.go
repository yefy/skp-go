package service

import (
	"fmt"
	"reflect"
)

type ObjMethod struct {
	name    string
	value   reflect.Value
	argvIn  int
	argvOut int
}

func CheckKind(obj interface{}, checkKind reflect.Kind) {
	objKind := reflect.TypeOf(obj).Kind()
	if objKind != checkKind {
		panic(fmt.Sprintf("checkKind = %+v, objKind = %+v \n", checkKind, objKind))
	}
}

func GetObjName(obj interface{}) string {
	objType := reflect.TypeOf(obj)
	objKind := objType.Kind()
	if objKind == reflect.Ptr {
		return objType.Elem().Name()

	} else if objKind == reflect.Struct {
		return objType.Name()
	}
	return ""
}

func GetMethod(obj interface{}) []*ObjMethod {
	objType := reflect.TypeOf(obj)
	objKind := objType.Kind()
	if objKind == reflect.Ptr {
		var mSlice []*ObjMethod
		objValue := reflect.ValueOf(obj)
		for i := 0; i < objType.NumMethod(); i++ {
			m := objType.Method(i)
			objM := &ObjMethod{}
			objM.name = m.Name
			objM.value = objValue.MethodByName(m.Name)
			objM.argvIn = objM.value.Type().NumIn()
			objM.argvOut = objM.value.Type().NumOut()
			fmt.Printf("objM = %+v \n", objM)
			mSlice = append(mSlice, objM)
		}
		return mSlice
	} else if objKind == reflect.Func {
		var mSlice []*ObjMethod
		objValue := reflect.ValueOf(obj)
		objM := &ObjMethod{}
		objM.name = ""
		objM.value = objValue
		objM.argvIn = objM.value.Type().NumIn()
		objM.argvOut = objM.value.Type().NumOut()
		fmt.Printf("objM = %+v \n", objM)
		mSlice = append(mSlice, objM)
		return mSlice
	}
	return nil
}

func GetArgvValue(argv ...interface{}) []reflect.Value {
	if len(argv) > 0 {
		argvValue := make([]reflect.Value, len(argv))
		for index, value := range argv {
			argvValue[index] = reflect.ValueOf(value)
		}
		return argvValue
	}
	return nil
}

func GetArgvsValue(argvs []interface{}) []reflect.Value {
	if len(argvs) > 0 {
		argvValue := make([]reflect.Value, len(argvs))
		for index, value := range argvs {
			argvValue[index] = reflect.ValueOf(value)
		}
		return argvValue
	}
	return nil
}

func GetInterfaceValue(values []reflect.Value) []reflect.Value {
	if len(values) > 0 {
		argvValue := make([]reflect.Value, len(values))
		for index, value := range values {
			argvValue[index] = reflect.ValueOf(value.Interface())
		}
		return argvValue
	}
	return nil
}
