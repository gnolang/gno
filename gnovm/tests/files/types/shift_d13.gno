package main

import "fmt"

// it's like assign
func foo(a uint64, b float32) {
	fmt.Printf("%T \n", a)
	println(a)
}

func main() {
	x := 11
	foo(1<<x, int(1)>>x)
}

// Error:
// main/files/types/shift_d13.gno:13:2: cannot use int as float32
