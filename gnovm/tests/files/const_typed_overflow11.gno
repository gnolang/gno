package main

// An overflowing typed constant is rejected as an array length (#6392).

const x int8 = 127

func main() {
	var arr [x + 1]int
	println(len(arr))
}

// Error:
// main/const_typed_overflow11.gno:8:11-16: x + 1 (constant 128 of type int8) overflows int8

// TypeCheckError:
// main/const_typed_overflow11.gno:8:11: x + 1 (constant 128 of type int8) overflows int8
