// A var that references itself is a genuine cycle.
package main

var A = A

func main() {
	println(A)
}

// Error:
// main/var42.gno:4:5-10: invalid recursive value: A -> A

// TypeCheckError:
// main/var42.gno:4:5: initialization cycle: A refers to itself
