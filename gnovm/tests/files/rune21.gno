package main

// An untyped rune constant beyond int32 cannot be an operand of a
// non-constant shift expression of type rune.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	x := big + ('a' << s)
	println(x)
}

// Error:
// main/rune21.gno:11:2-23: rune overflows target kind

// TypeCheckError:
// main/rune21.gno:11:7: big (untyped rune constant 1241243320321) overflows rune
