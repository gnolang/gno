package main

// one untyped const, one typed const
func main() {
	println("b" | "a")
}

// Error:
// main/or_b3.gno:5:10-19: operator | not defined on: StringKind

// TypeCheckError:
// main/or_b3.gno:5:10: invalid operation: operator | not defined on "b" (untyped string constant)
