package main

// In an interface context ('a' << s) + (1 << 40) is a rune, which 1<<40
// overflows.

var s uint

func main() {
	var x interface{} = ('a' << s) + (1 << 40)
	println(x)
}

// Error:
// main/rune30.gno:9:6-44: bigint overflows target kind

// TypeCheckError:
// main/rune30.gno:9:36: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/rune30.gno:9:35: (1 << 40) (untyped rune constant 1099511627776) overflows rune
