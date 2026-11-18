package main

// A negative untyped rune constant does not fit an unsigned integer.

func main() {
	var u uint64 = 'b' - 'c'
	println(u)
}

// Error:
// main/rune15.gno:6:6-26: rune underflows target kind

// TypeCheckError:
// main/rune15.gno:6:17: cannot use 'b' - 'c' (untyped rune constant -1) as uint64 value in variable declaration (overflows)
