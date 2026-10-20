package main

func main() {
	const a uint64 = 18446744073709551615
	println(int(a))
}

// Error:
// main/convert9.gno:5:10-16: constant 18446744073709551615 overflows IntKind
