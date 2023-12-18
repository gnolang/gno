package main

import "fmt"

type integer *int

// TODO: this is go outputs, should check println and fmt.Printf
// 0x0
// <nil>
// main.integer
func main() {
	println(integer(nil))
	fmt.Println(integer(nil))
	fmt.Printf("%T \n", integer(nil))
}

// Output:
// (nil main.integer)
// <nil>
// *int
