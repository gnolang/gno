package main

// gc accepts this and prints 1241243320418: the int64 context types the
// shift. Gno rejects an untyped rune constant beyond int32 next to any
// non-constant operand, whatever the context.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var y int64 = ('a' << s) + big
	println(y)
}

// Error:
// main/rune32.gno:12:16-32: rune overflows target kind
