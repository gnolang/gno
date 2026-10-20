package main

func main() {
	const a3 = int32(1) << 33
}

// Error:
// main/overflow9.gno:3:1-5:2: constant overflows

// TypeCheckError:
// main/overflow9.gno:4:13: int32(1) << 33 (constant 8589934592 of type int32) overflows int32
