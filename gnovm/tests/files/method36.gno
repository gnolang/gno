package main

import "fmt"

type B struct{}

func (B) Foo() string { return "Foo from B" }

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
