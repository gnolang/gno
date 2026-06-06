package main

func main() {
	const a uint64 = 18446744073709551615
	println(int8(a))
}

// Error:
// main/convert9a.gno:5:10-17: constant 18446744073709551615 overflows Int8Kind

// TypeCheckError:
// main/convert9a.gno:5:15: constant 18446744073709551615 overflows int8
