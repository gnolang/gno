package main

func main() {
	const b3 = uint32(1) << 33
}

// Error:
// main/files/overflow13.gno:3:1: constant overflows

// TypeCheckError:
// main/files/overflow13.gno:4:13: uint32(1) << 33 (constant 8589934592 of type uint32) overflows uint32
