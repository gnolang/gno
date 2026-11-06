package main

// An unsigned typed constant cannot go below zero (#6392).

const d uint8 = 0

func main() {
	println(d - 1)
}

// Error:
// main/const_typed_overflow2.gno:8:10-15: d - 1 (constant -1 of type uint8) overflows uint8

// TypeCheckError:
// main/const_typed_overflow2.gno:8:10: d - 1 (constant -1 of type uint8) overflows uint8
