package main

import "fmt"

type integer *int

func main() {
	println(integer(nil))
	fmt.Println(integer(nil))
	fmt.Printf("%T \n", integer(nil))
}

// Output:
// (nil main.integer)
// <nil>
// *int
