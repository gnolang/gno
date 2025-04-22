package main

func main() {
	const a4 = int64(1) << 65
}

// Error:
// main/files/overflow10.gno:3:1: constant overflows

// TypeCheckError:
// main/files/overflow10.gno:4:13: int64(1) << 65 (constant 36893488147419103232 of type int64) overflows int64
