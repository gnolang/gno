package main

// both typed(different) const
func main() {
	println(1 << 'a')
	println(1 >> 'a')
}

// Error:
// main/files/types/shift_a3.gno:5:2: bigint overflows target kind

// TypeCheckError:
// main/files/types/shift_a3.gno:5:10: cannot use 1 << 'a' (untyped int constant 158456325028528675187087900672) as int value in argument to built-in println (overflows)
