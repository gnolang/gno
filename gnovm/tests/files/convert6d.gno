package main

func main() {
	const a int = -1
	println(uint64(a))
}

// Error:
// main/convert6d.gno:5:10-19: constant -1 overflows Uint64Kind

// TypeCheckError:
// main/convert6d.gno:5:17: constant -1 overflows uint64
