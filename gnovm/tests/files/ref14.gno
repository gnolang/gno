package main

import "fmt"

// ref of an array element that is an interface type.

func main() {
	var a [3]interface{}
	a[0] = 42
	a[1] = "hello"
	p := &a[1]
	fmt.Printf("%T\n", p)
	*p = 99
	fmt.Println(a[1])
}

// Output:
// *interface {}
// 99
