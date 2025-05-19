package main

func main() {
	const b1 = uint8(1) << 8
}

// Error:
// main/overflow11.gno:3:1-5:2: constant overflows

// TypeCheckError:
// main/overflow11.gno:4:13: uint8(1) << 8 (constant 256 of type uint8) overflows uint8
