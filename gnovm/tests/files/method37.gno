package main

import "fmt"

type Foo struct {
	A int
}

func (f Foo) GetValue() int {
	return f.A
}

func main() {
	f1 := Foo{A: 1}
	f1GetValue := f1.GetValue
	f1.A = 2
	fmt.Println(f1GetValue())
	fmt.Println(f1.A)
}

// Output:
// 1
// 2
