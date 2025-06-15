package main

func main() {
	const b3 = uint32(1) << 33
}

// Error:
// main/overflow13.gno:3:1-5:2: constant overflows

// TypeCheckError:
// main/overflow13.gno:4:13: uint32(1) << 33 (constant 8589934592 of type uint32) overflows uint32
