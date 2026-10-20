package main

// As rune30, for a println argument.

var s uint

func main() {
	println(('a' << s) + (1 << 40))
}

// Error:
// main/rune31.gno:8:2-33: bigint overflows target kind

// TypeCheckError:
// main/rune31.gno:8:24: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/rune31.gno:8:23: (1 << 40) (untyped rune constant 1099511627776) overflows rune
