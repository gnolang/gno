package main

// 1. checkOrConvert is the core logic, to convert type(recursively if needed)
// 2. shift expr might remain untyped, until reach its use context, assign, func Call, type call
// 3. assign should be checked, LHS of assign has a type, use it to mutate RHS type if it's untyped;
//    it's not coerced
// 4. type call should be checked, it's coerced even RHS is typed.
// 5. func call should be checked, it's not coerced.

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
	// assign to type conform and interface
	r = 1 << x // 2 step. first binary, second call type

	r.foo()

	fmt.Printf("%T \n", r) // TODO: should output main.U64 rather than the underlying type
}

// Output:
// bar
// uint64
