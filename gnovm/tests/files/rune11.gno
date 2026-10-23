package main

// As rune8, defaulting to rune (gc rejects this too).

func main() {
	var r = '\U0010ffff' * '\U0010ffff'
	println(r)
}

// Error:
// main/rune11.gno:6:10-37: untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')

// TypeCheckError:
// main/rune11.gno:6:10: cannot use '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) as rune value in variable declaration (overflows)
