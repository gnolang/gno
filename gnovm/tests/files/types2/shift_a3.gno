package main

func main() {
	println(1 << 'a')
	println(1 >> 'a')
}

// Error:
// main/files/types2/shift_a3.gno:4:2: bigint overflows target kind
