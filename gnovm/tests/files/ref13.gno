package main

import "fmt"

// ref of a nil named interface.

type Stringer interface {
	String() string
}

func main() {
	var x Stringer
	p := &x
	fmt.Printf("%T\n", p)
}

// Output:
// *main.Stringer
