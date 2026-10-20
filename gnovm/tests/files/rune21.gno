package main

// ('a' << s) + (1 << 40) is a rune: the constant 1<<40 overflows it.

var s uint

func main() {
	x := ('a' << s) + (1 << 40)
	println(x)
}

// Error:
// main/rune21.gno:8:2-29: bigint overflows target kind

// TypeCheckError:
// main/rune21.gno:8:21: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/rune21.gno:8:20: (1 << 40) (untyped rune constant 1099511627776) overflows rune
