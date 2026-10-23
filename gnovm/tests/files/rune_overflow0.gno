package main

// Regression: a negative untyped rune constant does not fit uint64.
func main() {
	var u uint64 = 'a' - 'b'
	println(u)
}

// Output:
// 18446744073709551615

// TypeCheckError:
// main/rune_overflow0.gno:5:17: cannot use 'a' - 'b' (untyped rune constant -1) as uint64 value in variable declaration (overflows)
