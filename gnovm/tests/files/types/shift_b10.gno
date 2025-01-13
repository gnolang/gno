package main

import "fmt"

func main() {
	x := 2
	r := float32(int(1) + 1<<x + 1)
	println(r)
	fmt.Printf("%T\n", r)
}

// Output:
// 6
// float32
