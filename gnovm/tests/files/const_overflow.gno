package main

// A typed constant operation must be representable in its type
// (Go spec, Constant expressions): the fold is exact, not the machine's
// wrapping arithmetic.

const c int8 = 100

func main() {
	println(c * 2)
}

// Error:
// main/const_overflow.gno:10:10-15: c * 2 (constant 200 of type int8) overflows int8

// TypeCheckError:
// main/const_overflow.gno:10:10: c * 2 (constant 200 of type int8) overflows int8
