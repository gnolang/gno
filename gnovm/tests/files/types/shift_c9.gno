package main

import "fmt"

func main() {
	x := 2
	r := 1<<x + int(1)
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// int
