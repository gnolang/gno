package main

import "fmt"

// ref of a slice element that is an interface type.

func main() {
	s := []interface{}{1, "two", 3.0}
	p := &s[0]
	fmt.Printf("%T\n", p)
	*p = "replaced"
	fmt.Println(s[0])
}

// Output:
// *interface {}
// replaced
