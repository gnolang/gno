package main

// As rune8, through a constant declaration.

func main() {
	const c = '\U0010ffff' * '\U0010ffff'
	var x int64 = c
	println(x)
}

// Error:
// main/rune9.gno:6:12-39: untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
