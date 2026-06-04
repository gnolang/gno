package main

func main() {
	const a uint64 = 18446744073709551615
	println(int64(a))
}

// Error:
// main/convert9d.gno:5:10-18: constant 18446744073709551615 overflows Int64Kind

// TypeCheckError:
// main/convert9d.gno:5:16: constant 18446744073709551615 overflows int64
