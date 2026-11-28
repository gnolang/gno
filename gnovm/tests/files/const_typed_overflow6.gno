package main

// Shifting a negative typed constant past its type is rejected, not
// zeroed (#6392).

const y int8 = -128

func main() {
	println(y << 1)
}

// Error:
// main/const_typed_overflow6.gno:9:10-16: y << 1 (constant -256 of type int8) overflows int8

// TypeCheckError:
// main/const_typed_overflow6.gno:9:10: y << 1 (constant -256 of type int8) overflows int8
