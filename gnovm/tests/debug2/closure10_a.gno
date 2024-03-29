package main

import "fmt"

func bar() func() func() int {
	x := 1

	// First level of closure, modifies x
	return func() func() int {
		x++
		// Second level of closure, returns x
		return func() int {
			return x
		}
	}
}

func main() {
	f := bar() // f is the first-level closure
	g := f()   // g is the second-level closure, x is incremented here

	fmt.Println(g()) // prints the value of x after being modified by the first-level closure
}

// Output:
// 2
