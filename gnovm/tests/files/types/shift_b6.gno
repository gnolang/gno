package main

import "fmt"

type R interface {
	foo()
}

type U64 uint64

func (u64 U64) foo() {
	println("bar")
}

func bar(r R) {
	r.foo()
}

// case of expr in type call
func main() {
	x := 2
	var r R
	// assign to interface
	r = 1 << x // 2 step. first binary, second call type

	r.foo()

	fmt.Printf("%T \n", r) // TODO: should output main.U64 rather than the underlying type
}

// Error:
// method foo not found in type int
