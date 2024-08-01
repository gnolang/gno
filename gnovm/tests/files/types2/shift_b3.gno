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
	var r U64
	r = U64(1<<x + int(1)) // 2 step. first binary, second call type

	// passing params of interface type
	bar(r)

	fmt.Printf("%T \n", r) // TODO: should output main.U64 rather than the underlying type
}

// Output:
// bar
// uint64
