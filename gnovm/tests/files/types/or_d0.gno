package main

// both untyped const
// TODO: dec value representation, and this should happen in process stage!!!
func main() {
	println(1.0 | 1)
}

// Error:
// main/files/types/or_d0.gno:6:10: operator | not defined on: BigdecKind

// TypeCheckError:
// main/files/types/or_d0.gno:6:10: invalid operation: operator | not defined on 1.0 (untyped float constant 1)
