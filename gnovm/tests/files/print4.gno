package main

import "fmt"

type A struct {
	s string
}

func main() {
	var a *A

	fmt.Printf("%T \n", a)

	println(a)
}

// Output:
// *struct { s string }
// invalid-pointer