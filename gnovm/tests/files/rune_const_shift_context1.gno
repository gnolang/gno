package main

// A non-constant shift of an untyped rune constant with no context
// types the constant as rune, which 1241243320321 overflows.

const c = '\U0010ffff' * '\U0010ffff'

var s uint = 1

func main() {
	x := c << s
	println(x)
}

// Error:
// main/rune_const_shift_context1.gno:11:2-13: constant 1241243320321 overflows rune

// TypeCheckError:
// main/rune_const_shift_context1.gno:11:7: c (untyped rune constant 1241243320321) overflows rune
