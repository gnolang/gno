package main

func main() {
	const c1 = 1 << 128
	println(c1)
}

// Error:
// main/overflow16.gno:5:2-13: bigint overflows target kind
