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
	r := uint64(1<<x + int(1)) // 2 step. first binary, second call type
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// uint64
