package main

import "fmt"

// Reassign different concrete type through *Stringer pointer.

type Stringer interface {
	String() string
}

type A struct{}

func (A) String() string { return "A" }

type B struct{}

func (B) String() string { return "B" }

func main() {
	var x Stringer = A{}
	p := &x
	fmt.Println((*p).String())
	*p = B{}
	fmt.Println((*p).String())
}

// Output:
// A
// B
