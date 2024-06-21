package main

// one untyped const, one typed const
func main() {
	println(1 & "a")
}

// Error:
// main/files/types/and_b2.gno:5: operator & not defined on: StringKind
