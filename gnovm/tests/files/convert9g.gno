package main

func main() {
	const a uint64 = 18446744073709551615
	println(uint32(a))
}

// Error:
// main/convert9g.gno:5:10-19: constant 18446744073709551615 overflows Uint32Kind

// TypeCheckError:
// main/convert9g.gno:5:17: constant 18446744073709551615 overflows uint32
