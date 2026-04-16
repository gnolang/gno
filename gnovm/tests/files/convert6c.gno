package main

func main() {
	const a int = -1
	println(uint32(a))
}

// Error:
// main/convert6c.gno:5:10-19: constant -1 overflows Uint32Kind

// TypeCheckError:
// main/convert6c.gno:5:17: constant -1 overflows uint32
