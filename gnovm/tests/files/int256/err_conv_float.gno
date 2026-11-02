package main

func main() {
	println(uint256(1.5))
}

// Error:
// main/err_conv_float.gno:4:10-22: cannot convert (const (1.5 <untyped> bigdec)) to integer type

// TypeCheckError:
// main/err_conv_float.gno:4:18: cannot convert 1.5 (untyped float constant) to type uint256
