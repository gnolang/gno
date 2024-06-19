package main

// one untyped const, one typed const
func main() {
	println("b" | "a")
}

// Error:
// main/files/types/or_b3.gno:5: operator | not defined on: StringKind
