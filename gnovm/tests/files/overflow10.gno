package main

func main() {
	const a4 = int64(1) << 65
}

// Error:
// main/overflow10.gno:4:13-27: int64(1) << 65 (constant 36893488147419103232 of type int64) overflows int64

// TypeCheckError:
// main/overflow10.gno:4:13: int64(1) << 65 (constant 36893488147419103232 of type int64) overflows int64
