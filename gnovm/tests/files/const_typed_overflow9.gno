package main

// A constant of a named type is checked against that type (#6392).

type T int8

const x T = 127

func main() {
	println(x + 1)
}

// Error:
// main/const_typed_overflow9.gno:10:10-15: x + 1 (constant 128 of type main.T) overflows main.T

// TypeCheckError:
// main/const_typed_overflow9.gno:10:10: x + 1 (constant 128 of int8 type T) overflows int8
