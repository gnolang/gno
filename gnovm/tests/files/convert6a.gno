package main

func main() {
	const a int = -1
	println(uint8(a))
}

// Error:
// main/files/convert6a.gno:5:10: cannot convert constant of type IntKind to Uint8Kind

// TypeCheckError:
// main/files/convert6a.gno:5:16: constant -1 overflows uint8
