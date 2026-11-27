package main

func main() {
	const a uint8 = 1
	println(-a)
}

// Error:
// main/const_ovf_neg_uint8.gno:5:10-12: -a (constant -1 of type uint8) overflows uint8

// TypeCheckError:
// main/const_ovf_neg_uint8.gno:5:10: -a (constant -1 of type uint8) overflows uint8
