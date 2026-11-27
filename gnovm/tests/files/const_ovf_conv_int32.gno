package main

func main() {
	println(int32(int64(1) << 40))
}

// Error:
// main/const_ovf_conv_int32.gno:4:10-31: constant 1099511627776 overflows Int32Kind

// TypeCheckError:
// main/const_ovf_conv_int32.gno:4:16: constant 1099511627776 overflows int32
