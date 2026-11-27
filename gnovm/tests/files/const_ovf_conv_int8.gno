package main

func main() {
	println(int8(200))
}

// Error:
// main/const_ovf_conv_int8.gno:4:10-19: bigint overflows target kind

// TypeCheckError:
// main/const_ovf_conv_int8.gno:4:15: constant 200 overflows int8
