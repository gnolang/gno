package main

import "fmt"

func foo(a uint64, b float32) {
	fmt.Printf("%T \n", a)
	println(a)
}

func main() {
	x := 11
	foo(1<<x, 1>>x)
}

// Error:
// operators >> and >>= not defined for float32
