package main

// An untyped rune constant beyond int32 cannot default to rune.

func main() {
	var r = '\U0010ffff' * '\U0010ffff'
	println(r)
}

// Error:
// main/rune6.gno:6:6-37: rune overflows target kind

// TypeCheckError:
// main/rune6.gno:6:10: cannot use '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) as rune value in variable declaration (overflows)
