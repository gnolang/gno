package main

// An operation between an untyped rune constant beyond int32 and an untyped
// integer constant is an untyped rune, as in Go, and stays exact in
// constant contexts.

const big = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func main() {
	var x int64 = big + 1
	var y int64 = (big+0)*2 + 1
	const c = big - 4294967296 + 'a' // back in range: a rune
	r := c
	_ = []rune{r}
	println(x, y, r, big+0 == 4294967296, 0+big > 1<<31)
}

// Output:
// 4294967297 8589934593 97 true true
