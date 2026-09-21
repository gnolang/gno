// An untyped float constant that overflows the target type must be
// rejected at compile time, matching Go, not silently folded to +Inf.

package main

func main() {
	x := float32(1e39)
	println(x)
}

// Error:
// main/float11.gno:7:7-20: cannot convert untyped bigdec to float32 -- too close to +-Inf

// TypeCheckError:
// main/float11.gno:7:15: cannot convert 1e39 (untyped float constant 1e+39) to type float32
