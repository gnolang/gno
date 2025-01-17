package main

func main() {
	const a2 = int16(1) << 16
}

// Error:
// main/files/overflow8.gno:3:1: constant overflows
