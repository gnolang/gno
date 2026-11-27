package main

func main() {
	const a uint16 = 1
	println(a << 16)
}

// Error:
// main/const_ovf_shl_uint16.gno:5:10-17: a << 16 (constant 65536 of type uint16) overflows uint16

// TypeCheckError:
// main/const_ovf_shl_uint16.gno:5:10: a << 16 (constant 65536 of type uint16) overflows uint16
