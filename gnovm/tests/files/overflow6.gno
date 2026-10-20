package main

func main() {
	const a = int32(1) << 33
}

// Error:
// main/overflow6.gno:3:1-5:2: constant overflows
