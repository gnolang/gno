package main

// gc accepts this: the int64 conversion types both shifts, giving
// 420906795008 for s = 32. Gno does not resolve the type of two
// non-constant operands of different untyped kinds from their context,
// and rejects the expression rather than compute it in another type.

var s uint = 32

func main() {
	println(int64(('a' << s) + (1 << s)))
}

// Error:
// main/rune26.gno:11:16-37: incompatible types in binary expression: <untyped> rune ADD <untyped> bigint
