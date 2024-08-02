package main

import "fmt"

// case of expr in type call
func main() {
	x := 2
	r := float32(int(1) + 1<<x) // 2 step. first binary recursively, second call type
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// float32
