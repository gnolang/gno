package main

func main() {
	const c uint256 = 1 << 200
	println(float32(c))
}

// Error:
// main/err_conv_float32.gno:5:10-20: constant overflows Float32Kind

// TypeCheckError:
// main/err_conv_float32.gno:5:18: cannot convert c (constant 1606938044258990275541962092341162602522202993782792835301376 of uintptr type uint256) to type float32
