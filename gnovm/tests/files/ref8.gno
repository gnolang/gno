package main

import "fmt"

// ref of a named interface variable holding a concrete value.

type Stringer interface {
	String() string
}

type mystr struct{ s string }

func (m mystr) String() string { return m.s }

func main() {
	var x Stringer = mystr{"hello"}
	p := &x
	fmt.Printf("%T\n", p)
}

// Output:
// *main.Stringer
