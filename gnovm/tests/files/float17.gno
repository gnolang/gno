// A negative untyped float constant that overflows the target type must be
// rejected at compile time, like its positive counterpart.
package main

func main() {
	x := float32(-1e39)
	println(x)
}

// Error:
// main/float17.gno:6:7-21: cannot convert untyped bigdec to float32 -- too close to +-Inf

// TypeCheckError:
// main/float17.gno:6:15: cannot convert -1e39 (untyped float constant -1e+39) to type float32
