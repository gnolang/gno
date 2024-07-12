package main

// one untyped const, one typed const
func main() {
	println("b" & "a")
}

// Error:
// main/files/types/and_b3.gno:5:10: operator & not defined on: StringKind
