package main

// In an interface context, the untyped rune constant beyond int32 takes
// its default type, rune, which it overflows.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var x interface{} = ('a' << s) + big
	println(x)
}

// Error:
// main/rune7.gno:11:6-38: rune overflows target kind

// TypeCheckError:
// main/rune7.gno:11:35: big (untyped rune constant 1241243320321) overflows rune
