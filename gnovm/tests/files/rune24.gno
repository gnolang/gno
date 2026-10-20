package main

// The non-constant switch tag defaults to rune before the cases are
// typed: the case constant overflows it.

var s uint

func main() {
	switch '\x00' << s {
	case '\x01' << 40:
		println("bad")
	}
}

// Error:
// main/rune24.gno:10:2-11:17: rune overflows target kind

// TypeCheckError:
// main/rune24.gno:10:7: '\x01' << 40 (untyped rune constant 1099511627776) overflows rune
