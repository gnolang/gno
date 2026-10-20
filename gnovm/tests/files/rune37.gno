package main

// As rune35, for 0 + big on the left and <.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println((0 + big) < (1 << s))
}

// Error:
// main/rune37.gno:10:10-30: rune overflows target kind

// TypeCheckError:
// main/rune37.gno:10:11: 0 + big (untyped rune constant 4294967296) overflows rune; main/rune37.gno:10:10: (0 + big) (untyped rune constant 4294967296) overflows rune
