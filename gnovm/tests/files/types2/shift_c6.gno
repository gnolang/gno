package main

// 1. checkOrConvert is the core logic, to convert type(recursively if needed)
// 2. shift expr might remain untyped, until reach its use context, assign, func Call, type call
// 3. assign should be checked, LHS of assign has a type, use it to mutate RHS type if it's untyped;
//    it's not coerced
// 4. type call should be checked, it's coerced even RHS is typed.
// 5. func call should be checked, it's not coerced.

import "fmt"

// case of expr in type call
func main() {
	x := 2
	var r float32
	r = 1 << x // untyped rhs to float32
	println(r)
	fmt.Printf("%T \n", r)
}

// Error:
// main/files/types2/shift_c6.gno:16: operator << not defined on: Float32Kind
