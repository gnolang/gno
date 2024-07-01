package main

import "fmt"

func foo(a uint64, b int64) {
	fmt.Printf("%T \n", a)
	fmt.Printf("%T \n", b)
	println(a)
	println(b)
}

func main() {
	x := 11
	foo(1<<x, 1>>x)
}

// Output:
// uint64
// int64
// 2048
// 0
