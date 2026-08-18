package main

// Tests transitive dependency resolution through a chain of functions:
// A = f() depends on g() which depends on B.
// B must be initialized before A even though A is declared first.

func f() int { return g() }
func g() int { return B }

var (
	A = f()
	B = 42
)

func main() { println(A) }

// Output:
// 42
