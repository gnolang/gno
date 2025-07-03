package main

import "fmt"

// var f func()
var a *struct{}
var b any
var c map[string]int
var s []int

func main() {
	if nil == a {
		fmt.Println("pointer == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if nil == b {
		fmt.Println("interface == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if nil == c {
		fmt.Println("map == nil")
	} else {
		println("not nil!")
	}
	fmt.Println("----")

	if nil == s {
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
