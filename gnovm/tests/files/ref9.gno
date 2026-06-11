package main

import "fmt"

// ref of a struct field that is an interface type.

type S struct {
	F interface{}
}

func main() {
	s := S{F: 42}
	p := &s.F
	fmt.Printf("%T\n", p)
	*p = "hello"
	fmt.Println(s.F)
}

// Output:
// *interface {}
// hello
