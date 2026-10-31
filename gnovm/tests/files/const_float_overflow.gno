package main

// A typed float constant operation must be representable in its type
// (#6392): the exact product is rounded once, and infinity is not a value.

const big float64 = 1e300

func main() {
	println(big * big)
}

// Error:
// main/const_float_overflow.gno:9:10-19: big * big (constant 1e+600 of type float64) overflows float64

// TypeCheckError:
// main/const_float_overflow.gno:9:10: big * big (constant 1e+600 of type float64) overflows float64
