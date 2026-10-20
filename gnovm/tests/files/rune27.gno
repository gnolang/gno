package main

// gc accepts this (420906795008 for s = 32); Gno rejects it, see rune26.

var s uint = 32

func main() {
	var x int64 = ('a' << s) + (1 << s)
	println(x)
}

// Error:
// main/rune27.gno:8:16-37: incompatible types in binary expression: <untyped> rune ADD <untyped> bigint
