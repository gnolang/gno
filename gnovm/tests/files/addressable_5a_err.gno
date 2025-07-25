package main

import "math"

func main() {
	// Untyped constants are not addressable.
	_ = &math.MaxUint8
}

// Error:
// illegal assignment X expression type *gnolang.ConstExpr

// TypeCheckError:
// main/addressable_5a_err.gno:7:7: invalid operation: cannot take address of math.MaxUint8 (untyped int constant 255)
