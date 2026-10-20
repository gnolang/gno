package main

func main() {
	const a2 = int16(1) << 16
}

// Error:
// main/overflow8.gno:3:1-5:2: constant overflows
