package main

func main() {
	println(uint8(int8(-1)))
}

// Error:
// main/const_ovf_conv_neg_uint8.gno:4:10-25: constant -1 overflows Uint8Kind

// TypeCheckError:
// main/const_ovf_conv_neg_uint8.gno:4:16: constant -1 overflows uint8
