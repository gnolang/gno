package main

func main() {
	const b2 = uint16(1) << 16
}

// Error:
// main/overflow12.gno:3:1-5:2: constant overflows
