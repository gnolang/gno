package main

// A negative overflowing shift (gc: -4294967296). Master's shift check
// tested only the upper bound, so it wrapped this and printed 0.

func main() {
	var x int64 = ('\x00' - '\x01') << 32
	println(x)
}

// Error:
// main/rune24.gno:6:1-9:2: untyped rune constant ('\x00' - '\x01') << 32 overflows int32 (value -4294967296); convert an operand to a wider integer type, e.g. int64('\x00' - '\x01')
