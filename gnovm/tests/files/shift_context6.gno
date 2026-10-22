package main

func main() {
	var s uint = 32
	println(('a' << s) + (1 << 40))
}

// Error:
// main/shift_context6.gno:5:2-33: bigint overflows target kind

// TypeCheckError:
// main/shift_context6.gno:5:24: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/shift_context6.gno:5:23: (1 << 40) (untyped rune constant 1099511627776) overflows rune
