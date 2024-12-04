package main

import (
	"fmt"
	"reflect"
)

func Add(a, b int) int {
	return a + b
}

func main() {
	// 함수 호출
	result := executeFunction(reflect.ValueOf(Add), 3, 5)
	fmt.Println("Result:", result[0].Interface())
}

func executeFunction(fn reflect.Value, args ...interface{}) []reflect.Value {
	// 인수를 reflect.Value로 변환
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 함수 호출
	return fn.Call(in)
}
