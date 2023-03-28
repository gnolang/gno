package main

import (
	"fmt"
	"reflect"
	"time"
)

type MyWriter interface {
	Write(p []byte) (i int, err error)
}

type TestStruct struct{}

func (t TestStruct) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func usesWriter(w MyWriter) {
	n, _ := w.Write([]byte("hello world"))
	fmt.Println(n)
}

type MyStringer interface {
	String() string
}

func usesStringer(s MyStringer) {
	fmt.Println(s.String())
}

func main() {
	aType := reflect.TypeOf((*MyWriter)(nil)).Elem()

	var t interface{}
	t = TestStruct{}
	var tw MyWriter
	var ok bool
	tw, ok = t.(MyWriter)
	if !ok {
		fmt.Println("TestStruct does not implement MyWriter")
	} else {
		fmt.Println("TestStruct implements MyWriter")
		usesWriter(tw)
	}
	n, _ := t.(MyWriter).Write([]byte("hello world"))
	fmt.Println(n)
	bType := reflect.TypeOf(TestStruct{})
	fmt.Println(aType.IsImplementedBy(bType))

	t = 42
	foo, ok := t.(MyWriter)
	if !ok {
		fmt.Println("42 does not implement MyWriter")
	} else {
		fmt.Println("42 implements MyWriter")
	}
	_ = foo

	var tt interface{}
	tt = time.Nanosecond
	var myD MyStringer
	myD, ok = tt.(MyStringer)
	if !ok {
		fmt.Println("time.Nanosecond does not implement MyStringer")
	} else {
		fmt.Println("time.Nanosecond implements MyStringer")
		usesStringer(myD)
	}
	fmt.Println(tt.(MyStringer).String())
	cType := reflect.TypeOf((*MyStringer)(nil)).Elem()
	dType := reflect.TypeOf(time.Nanosecond)
	fmt.Println(cType.IsImplementedBy(dType))

	tt = 42
	bar, ok := tt.(MyStringer)
	if !ok {
		fmt.Println("42 does not implement MyStringer")
	} else {
		fmt.Println("42 implements MyStringer")
	}
	_ = bar

}

// Output:
// TestStruct implements MyWriter
// 11
// 11
// true
// 42 does not implement MyWriter
// time.Nanosecond implements MyStringer
// 1ns
// 1ns
// true
// 42 does not implement MyStringer
