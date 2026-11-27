package main

func main() {
	const f uint64 = 1 << 63
	println(f + f)
}

// Error:
// main/const_ovf_add_uint64.gno:5:10-15: f + f (constant 18446744073709551616 of type uint64) overflows uint64

// TypeCheckError:
// main/const_ovf_add_uint64.gno:5:10: f + f (constant 18446744073709551616 of type uint64) overflows uint64
