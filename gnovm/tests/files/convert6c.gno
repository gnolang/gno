package main

func main() {
	const a int = -1
	println(uint32(a))
}

// Error:
// main/convert6c.gno:5:10-19: cannot convert constant of type IntKind to Uint32Kind

// TypeCheckError:
// main/convert6c.gno:5:17: constant -1 overflows uint32
