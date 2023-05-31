package main

import "fmt"

func main() {
	var x interface{}
	fmt.Printf("%+v \n", x)
	println(x)
	println(x == nil)
}

// Output:
// <nil>
// nil
// true
