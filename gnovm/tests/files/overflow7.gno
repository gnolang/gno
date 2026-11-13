package main

func main() {
	const a1 = int8(1) << 8
}

// Error:
// main/overflow7.gno:4:13-25: int8(1) << 8 (constant 256 of type int8) overflows int8

// TypeCheckError:
// main/overflow7.gno:4:13: int8(1) << 8 (constant 256 of type int8) overflows int8
