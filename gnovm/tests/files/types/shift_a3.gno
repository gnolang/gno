package main

// both typed(different) const
func main() {
	println(1 << 'a')
	println(1 >> 'a')
}

// Error:
// main/shift_a3.gno:5:2-19: bigint overflows target kind
