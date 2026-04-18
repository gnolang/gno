package main

import "fmt"

// Closure capturing &i where i is interface{}.
// Exercises heap-use path in PopAsPointer2.

func main() {
	var i interface{} = 42
	f := func() {
		p := &i
		fmt.Printf("%T\n", p)
		*p = "hello"
	}
	f()
	fmt.Println(i)
}

// Output:
// *interface {}
// hello
