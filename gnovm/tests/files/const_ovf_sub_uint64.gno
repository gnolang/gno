package main

func main() {
	println(uint64(8) - uint64(40))
}

// Error:
// main/const_ovf_sub_uint64.gno:4:10-32: uint64(8) - uint64(40) (constant -32 of type uint64) overflows uint64

// TypeCheckError:
// main/const_ovf_sub_uint64.gno:4:10: uint64(8) - uint64(40) (constant -32 of type uint64) overflows uint64
