package main

// A negative untyped rune constant does not fit an unsigned integer.

func main() {
	var u uint64 = -'a'
	println(u)
}

// Error:
// main/rune13.gno:6:6-21: rune underflows target kind

// TypeCheckError:
// main/rune13.gno:6:17: cannot use -'a' (untyped rune constant -97) as uint64 value in variable declaration (overflows)
