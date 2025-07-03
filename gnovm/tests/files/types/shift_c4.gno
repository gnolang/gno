package main

import "fmt"

func main() {
	x := 63
	r := int32(+(1<<x + int64(1)))
	println(r)
	fmt.Printf("%T\n", r)
}

// Output:
// 1
// int32
