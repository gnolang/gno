package main

func main() {
	const c1 = 1 << 128
	println(c1)
}

// Error:
// main/files/overflow16.gno:5:2: bigint overflows target kind
