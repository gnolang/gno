package main

// As rune35, for big * 1 on the left and !=.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println((big * 1) != (1 << s))
}

// Error:
// main/rune36.gno:10:10-31: rune overflows target kind

// TypeCheckError:
// main/rune36.gno:10:11: big * 1 (untyped rune constant 4294967296) overflows rune; main/rune36.gno:10:10: (big * 1) (untyped rune constant 4294967296) overflows rune
