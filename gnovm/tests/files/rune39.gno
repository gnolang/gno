package main

// As rune35, through a named constant.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	const c = big + 0
	println((1 << s) == c)
}

// Error:
// main/rune39.gno:11:10-23: rune overflows target kind

// TypeCheckError:
// main/rune39.gno:11:22: c (untyped rune constant 4294967296) overflows rune
