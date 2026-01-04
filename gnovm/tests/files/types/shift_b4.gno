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

func main() {
	x := 2
	var r R
	// assign to interface
	r = U64(1<<x + int(1))

	r.foo()

	fmt.Printf("%T\n", r) // TODO: should output main.U64 rather than the underlying type
}

// Output:
// bar
// main.U64
