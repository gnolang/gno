package main

// big + 0 is an untyped rune constant, like big: it overflows rune next to 1<<s.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println((1 << s) == (big + 0))
}

// Error:
// main/rune35.gno:10:10-31: rune overflows target kind

// TypeCheckError:
// main/rune35.gno:10:23: big + 0 (untyped rune constant 4294967296) overflows rune; main/rune35.gno:10:22: (big + 0) (untyped rune constant 4294967296) overflows rune
