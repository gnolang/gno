package main

// A large untyped rune shifted by a variable, as a switch tag, overflows
// its default type rune.

var s uint

func main() {
	switch '\x01' << 40 << s {
	case 0:
		println("bad")
	}
}

// Error:
// main/rune25.gno:9:2-12:3: rune overflows target kind

// TypeCheckError:
// main/rune25.gno:9:9: '\x01' << 40 (untyped rune constant 1099511627776) overflows rune
