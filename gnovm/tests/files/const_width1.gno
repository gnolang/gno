package main

// An untyped integer constant is at most 512 bits, as in go/types (prec):
// a wider shift result is rejected before it can be used.

const c = 1 << 512

func main() {
	println(c > 0)
}

// Error:
// main/const_width1.gno:6:11-19: constant shift overflow

// TypeCheckError:
// main/const_width1.gno:6:13: constant shift overflow
