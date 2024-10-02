package main

// both typed(different) const
func main() {
	println(1 << 'a')
	println(1 >> 'a')
}

// Error:
// main/files/types/shift_a3.gno:5:2: bigint overflows target kind
