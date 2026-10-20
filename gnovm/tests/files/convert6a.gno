package main

func main() {
	const a int = -1
	println(uint8(a))
}

// Error:
// main/convert6a.gno:5:10-18: constant -1 overflows Uint8Kind
