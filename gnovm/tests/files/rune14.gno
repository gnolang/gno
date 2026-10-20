package main

// As rune8, for unary minus: -(-1 << 31) is 2147483648.

func main() {
	var x int64 = -(-'\x01' << 31)
	println(x)
}

// Error:
// main/rune14.gno:6:16-32: untyped rune constant -((-'\x01') << 31) overflows int32 (value 2147483648); convert an operand to a wider integer type, e.g. int64((-'\x01') << 31)
