package main

// both untyped const
// TODO: dec value representation
func main() {
	println(1.0 | 0)
}

// Error:
// main/files/types/or_d4.gno:6:10: operator | not defined on: BigdecKind
