package main

func main() {
	const a int = -1
	println(uint16(a))
}

// Error:
// main/files/convert6b.gno:5:10: cannot convert constant of type IntKind to Uint16Kind

// TypeCheckError:
// main/files/convert6b.gno:5:17: constant -1 overflows uint16
