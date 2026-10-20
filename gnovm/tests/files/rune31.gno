package main

// As rune30, with the operands swapped and !=.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println(big != (1 << s))
}

// Error:
// main/rune31.gno:10:10-25: rune overflows target kind

// TypeCheckError:
// main/rune31.gno:10:10: big (untyped rune constant 4294967296) overflows rune
