package main

// A unary switch tag has the type of its operand, rune, which the
// non-constant case overflows.

var s uint

func main() {
	switch +('\x00' << s) {
	case ('\x01' << 40) << s:
		println("bad")
	}
}

// Error:
// main/rune26.gno:10:7-26: rune overflows target kind

// TypeCheckError:
// main/rune26.gno:10:8: '\x01' << 40 (untyped rune constant 1099511627776) overflows rune; main/rune26.gno:10:7: ('\x01' << 40) (untyped rune constant 1099511627776) overflows rune
