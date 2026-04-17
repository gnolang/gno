package main

import "fmt"

// ref of an interface{} variable, then assign a different type through the pointer.
// In Go this is valid: *p = "hello" changes i from int to string.
// If the bug makes &i produce *int, then assigning a string would fail.

func main() {
	var i interface{} = 42
	p := &i
	*p = "hello"
	fmt.Println(i)
}

// Output:
// hello
