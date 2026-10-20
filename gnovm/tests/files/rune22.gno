package main

// The shifted constant takes its default type, rune, which it overflows.

var s uint

func main() {
	x := ('\U0010ffff' * '\U0010ffff') << s
	println(x)
}

// Error:
// main/rune22.gno:8:7-41: rune overflows target kind

// TypeCheckError:
// main/rune22.gno:8:8: '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) overflows rune; main/rune22.gno:8:7: ('\U0010ffff' * '\U0010ffff') (untyped rune constant 1241243320321) overflows rune
