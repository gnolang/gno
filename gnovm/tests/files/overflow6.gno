package main

func main() {
	const a = int32(1) << 33
}

// Error:
// main/files/overflow6.gno:3:1: constant overflows

// TypeCheckError:
// main/files/overflow6.gno:4:12: int32(1) << 33 (constant 8589934592 of type int32) overflows int32
