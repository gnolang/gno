package main

// both typed(different) const
func main() {
	println(1 << 'a')
	println(1 >> 'a')
}

// Error:
// main/files/types/10a3_filetest.gno:5: bigint overflows target kind
