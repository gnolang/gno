package main

import "fmt"

type X struct {
	v string
}

func main() {
	x := X{v: "a"}
	fmt.Printf("test %#v\n", x)
}

// Output:
// test struct { v string }{v:"a"}
