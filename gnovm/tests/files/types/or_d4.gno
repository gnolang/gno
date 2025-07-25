package main

// both untyped const
// TODO: dec value representation
func main() {
	println(1.0 | 0)
}

// Error:
// main/or_d4.gno:6:10-17: operator | not defined on: BigdecKind

// TypeCheckError:
// main/or_d4.gno:6:10: invalid operation: operator | not defined on 1.0 (untyped float constant 1)
