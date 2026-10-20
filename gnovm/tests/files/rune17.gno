package main

// A case constant beyond int32 does not fit a non-constant rune switch
// tag.

var s uint

func main() {
	switch '\x00' << s {
	case '\x01' << 40:
		println("bad")
	}
}

// Error:
// main/rune17.gno:10:2-11:17: bigint overflows target kind

// TypeCheckError:
// main/rune17.gno:10:7: '\x01' << 40 (untyped rune constant 1099511627776) overflows rune
