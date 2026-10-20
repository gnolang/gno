package main

// As rune35, for nested constant folds.

var s uint = 32

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	println((1 << s) > ((big+0)*1 - 0))
}

// Error:
// main/rune38.gno:10:10-36: rune overflows target kind

// TypeCheckError:
// main/rune38.gno:10:22: (big + 0) * 1 - 0 (untyped rune constant 4294967296) overflows rune; main/rune38.gno:10:21: ((big + 0) * 1 - 0) (untyped rune constant 4294967296) overflows rune
