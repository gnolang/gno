package main

func main() {
	const a int = -1
	println(uint(a))
}

// Error:
// main/convert6.gno:5:10-17: cannot convert constant of type IntKind to UintKind

// TypeCheckError:
// main/convert6.gno:5:15: constant -1 overflows uint
