package main

func main() {
	const a int = -1
	println(uint(a))
}

// Error:
// main/files/convert6.gno:5:10: cannot convert constant of type IntKind to UintKind

// TypeCheckError:
// main/files/convert6.gno:5:15: constant -1 overflows uint
