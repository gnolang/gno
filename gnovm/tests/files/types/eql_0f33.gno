package main

import "fmt"

// var f func()
var a *struct{}
var b any
var c map[string]int
var s []int

func main() {
	if a == nil {
		fmt.Println("pointer == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if b == nil {
		fmt.Println("interface == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if c == nil {
		fmt.Println("map == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if s == nil {
		fmt.Println("slice == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")
}

// Output:
// pointer == nil
// ----
// interface == nil
// ----
// map == nil
// ----
// slice == nil
// ----
