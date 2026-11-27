package main

func main() {
	const a uint64 = 1
	println(a << 64)
}

// Error:
// main/const_ovf_shl_uint64.gno:5:10-17: a << 64 (constant 18446744073709551616 of type uint64) overflows uint64

// TypeCheckError:
// main/const_ovf_shl_uint64.gno:5:10: a << 64 (constant 18446744073709551616 of type uint64) overflows uint64
