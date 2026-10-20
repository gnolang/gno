package main

// Two levels down, on the right of the comparison.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	println('a' > ((('a' << s) + big) * 2))
}

// Error:
// main/rune28.gno:10:18-34: rune overflows target kind

// TypeCheckError:
// main/rune28.gno:10:31: big (untyped rune constant 1241243320321) overflows rune
