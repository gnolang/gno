package main

// A negative untyped rune constant does not fit an unsigned type (#6389).

func main() {
	var u uint64 = 'b' - 'c'
	println(u)
}

// Error:
// main/rune_const_overflow2.gno:6:6-26: constant -1 overflows uint64

// TypeCheckError:
// main/rune_const_overflow2.gno:6:17: cannot use 'b' - 'c' (untyped rune constant -1) as uint64 value in variable declaration (overflows)
