package main

func main() {
	const a uint64 = 18446744073709551615
	println(uint8(a))
}

// Error:
// main/convert9e.gno:5:10-18: constant 18446744073709551615 overflows Uint8Kind

// TypeCheckError:
// main/convert9e.gno:5:16: constant 18446744073709551615 overflows uint8
