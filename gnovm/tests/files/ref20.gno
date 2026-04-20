package main

import "fmt"

// Deref of *Stringer preserves concrete type.

type Stringer interface {
	String() string
}

type mystr struct{ s string }

func (m mystr) String() string { return m.s }

func main() {
	var x Stringer = mystr{"hello"}
	p := &x
	v := *p
	fmt.Println(v.String())
}

// Output:
// hello
