// MAXALLOC: 1000000
package main

type Big [300000]byte

func f() Big { return Big{} }

var x, y any

// Regression: doOpAssign popped its Rhs values before assigning, so the
// fresh values of `f(), f()`, reached only through the value stack while
// each was copied into its target, were invisible to GC.
func main() {
	x, y = f(), f()
	println("survived")
}

// Error:
// allocation limit exceeded
