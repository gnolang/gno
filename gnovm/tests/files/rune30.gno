package main

// Mixed equality: 1<<s takes the type rune from big, which overflows it.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println((1 << s) == big)
}

// Error:
// main/rune30.gno:10:10-25: rune overflows target kind

// TypeCheckError:
// main/rune30.gno:10:22: big (untyped rune constant 4294967296) overflows rune
