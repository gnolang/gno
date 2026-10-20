package main

// Regression: the untyped rune product overflows rune, its default type.
func main() {
	var r = '\U0010ffff' * '\U0010ffff'
	println(r)
}

// Output:
// -2228223

// TypeCheckError:
// main/rune_overflow1.gno:5:10: cannot use '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) as rune value in variable declaration (overflows)
