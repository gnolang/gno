// A typed float constant that overflows the narrower target type must be
// rejected at compile time for either sign, matching Go.

package main

const big float64 = -3.5e38

func main() {
	x := float32(big)
	println(x)
}

// Error:
// main/float13.gno:9:7-19: cannot convert constant of type Float64Kind to Float32Kind

// TypeCheckError:
// main/float13.gno:9:15: cannot convert big (constant -3.5e+38 of type float64) to type float32
