package main

// As rune17, for a compound tag.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	switch ('\x00' << s) + '\x00' {
	case big:
		println("bad")
	default:
		println("miss")
	}
}

// Error:
// main/rune19.gno:11:2-12:17: bigint overflows target kind

// TypeCheckError:
// main/rune19.gno:11:7: big (untyped rune constant 1241243320321) overflows rune
