// An untyped integer constant that overflows the float target is rejected.
package main

func main() {
	x := float32(1 << 128)
	println(x)
}

// Error:
// main/float16.gno:5:7-24: bigint overflows float32

// TypeCheckError:
// main/float16.gno:5:15: cannot convert 1 << 128 (untyped int constant 340282366920938463463374607431768211456) to type float32
