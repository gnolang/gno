package main

func main() {
	const a4 = int64(1) << 65
}

// Error:
// main/files/overflow10.gno:3:1: constant overflows
