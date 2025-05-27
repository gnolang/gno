package main

func main() {
	const a int = -1
	println(uint16(a))
}

// Error:
// main/convert6b.gno:5:10-19: cannot convert constant of type IntKind to Uint16Kind

// TypeCheckError:
// main/convert6b.gno:5:17: constant -1 overflows uint16
