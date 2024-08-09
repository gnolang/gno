package main

import "fmt"

func main() {
	x := 2
	r := float32(int(1) + 1<<x)
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// float32
