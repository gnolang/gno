package main

// As rune17, for a unary tag and a non-constant case.

var s uint

func main() {
	switch +('\x00' << s) {
	case ('\x01' << 40) << s:
		println("bad")
	}
}

// Error:
// main/rune18.gno:9:2-10:17: bigint overflows target kind

// TypeCheckError:
// main/rune18.gno:9:8: '\x01' << 40 (untyped rune constant 1099511627776) overflows rune; main/rune18.gno:9:7: ('\x01' << 40) (untyped rune constant 1099511627776) overflows rune
