package main

// In a comparison with a rune shift, 1<<40 is a rune and overflows.

var s uint

func main() {
	println(('a' << s) > 1<<40)
}

// Error:
// main/rune23.gno:8:10-28: bigint overflows target kind

// TypeCheckError:
// main/rune23.gno:8:23: 1 << 40 (untyped rune constant 1099511627776) overflows rune
