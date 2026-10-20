package main

// A comparison with a non-constant rune shift gives the constant the type
// rune, which it overflows.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	println(('a' << s) < big)
}

// Error:
// main/rune23.gno:11:10-26: rune overflows target kind

// TypeCheckError:
// main/rune23.gno:11:23: big (untyped rune constant 1241243320321) overflows rune
