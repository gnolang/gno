package main

func main() {
	const a int = -1
	println(uint16(a))
}

// Error:
// main/convert6b.gno:5:10-19: constant -1 overflows Uint16Kind

// TypeCheckError:
// main/convert6b.gno:5:17: constant -1 overflows uint16
