package main

func main() {
	const a = int32(1) << 33
}

// Error:
// main/files/overflow6.gno:3:1: constant overflows
