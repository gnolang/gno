package main

// A constant shift count is at most 1074, as in go/types (shiftBound),
// even when the shifted value is zero.

func main() {
	println(0 << 1075)
}

// Error:
// main/const_width4.gno:7:10-19: invalid operation: invalid shift count 1075

// TypeCheckError:
// main/const_width4.gno:7:15: invalid operation: invalid shift count 1075 (untyped int constant)
