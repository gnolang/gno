package main

import "fmt"

// ref of an interface{} variable holding a concrete value.
// In Go, &i where i is interface{} always gives *interface{}.

func main() {
	var i interface{} = 42
	p := &i
	fmt.Printf("%T\n", p)
}

// Output:
// *interface {}
