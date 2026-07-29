// A typed float constant that overflows the narrower target type must be
// rejected at compile time, matching Go, not silently folded to +Inf.

package main

const big float64 = 1e39

func main() {
	x := float32(big)
	println(x)
}

// Error:
// main/float12.gno:9:7-19: cannot convert constant of type Float64Kind to Float32Kind

// TypeCheckError:
// main/float12.gno:9:15: cannot convert big (constant 1e+39 of type float64) to type float32
