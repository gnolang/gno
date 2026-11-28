package main

// An overflowing typed constant sum is rejected as an initializer too
// (#6392).

const x int8 = 100

func main() {
	var y int8 = x + x
	println(y)
}

// Error:
// main/const_typed_overflow10.gno:9:15-20: x + x (constant 200 of type int8) overflows int8

// TypeCheckError:
// main/const_typed_overflow10.gno:9:15: x + x (constant 200 of type int8) overflows int8
