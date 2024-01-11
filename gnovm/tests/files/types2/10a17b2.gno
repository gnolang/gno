package main

import "fmt"

// it's like assign
func foo(a uint64, b int64) {
	fmt.Printf("%T \n", a)
	println(a)
}

func main() {
	x := 11
	foo(1<<x, 1>>x)
}

// Output:
// uint64
// 2048
