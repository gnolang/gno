package main

import "fmt"

// case of expr in type call
func main() {
	x := 2
	r := uint64(1<<x + 1) // 2 step. first binary, second call type
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// uint64
