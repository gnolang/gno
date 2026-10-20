package main

// A nested operand: big meets a non-constant shift inside the comparison.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	println((big + ('a' << s)) < 'a')
}

// Error:
// main/rune27.gno:10:11-27: rune overflows target kind

// TypeCheckError:
// main/rune27.gno:10:11: big (untyped rune constant 1241243320321) overflows rune
