package main

func main() {
	const a4 = int64(1) << 65
}

// Error:
// main/overflow10.gno:3:1-5:2: constant overflows
