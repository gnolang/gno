package main

// The overflow diagnostic names constants as written.

const A = '\U0010ffff'

const B = A * A

func main() {
	println(B)
}

// Error:
// main/rune20.gno:7:11-16: untyped rune constant A * A overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64(A)

// TypeCheckError:
// main/rune20.gno:10:10: cannot use B (untyped rune constant 1241243320321) as rune value in argument to built-in println (overflows)
