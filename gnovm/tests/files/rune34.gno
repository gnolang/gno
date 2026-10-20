package main

// gc accepts this and prints 1241243320321; Gno rejects it (see rune32):
// the non-constant operand is a typed variable.

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var x int64
	println(x + big)
}

// Error:
// main/rune34.gno:10:10-17: rune overflows target kind
