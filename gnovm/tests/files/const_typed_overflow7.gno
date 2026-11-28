package main

// A shift that moves every bit out of a negative typed constant is
// rejected (#6392).

const a int8 = -1

func main() {
	println(a << 8)
}

// Error:
// main/const_typed_overflow7.gno:9:10-16: a << 8 (constant -256 of type int8) overflows int8

// TypeCheckError:
// main/const_typed_overflow7.gno:9:10: a << 8 (constant -256 of type int8) overflows int8
