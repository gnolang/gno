package main

// gc rejects this too: with the float constant the shifts would be float64.

var s uint

func main() {
	x := ('a' << s) + (1 << s) + 0.0
	println(x)
}

// Error:
// main/rune28.gno:8:7-28: incompatible types in binary expression: <untyped> rune ADD <untyped> bigint

// TypeCheckError:
// main/rune28.gno:8:8: invalid operation: shifted operand 'a' (type float64) must be integer; main/rune28.gno:8:21: invalid operation: shifted operand 1 (type float64) must be integer
