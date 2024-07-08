package main

import "fmt"

type Foo struct {
	A int
}

func (f Foo) WithValue(x int) *Foo {
	f.A = x
	return &f
}

func main() {
	f1 := Foo{A: 1}
	fmt.Println(f1.WithValue(2))
	fmt.Println(f1)
}

// Output:
// &{2}
// {1}
