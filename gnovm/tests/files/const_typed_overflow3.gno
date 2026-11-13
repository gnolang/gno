package main

// A typed int64 constant product past the type is rejected, not zeroed
// (#6392).

const e int64 = 1 << 62

func main() {
	println(e * 4)
}

// Error:
// main/const_typed_overflow3.gno:9:10-15: e * 4 (constant 18446744073709551616 of type int64) overflows int64

// TypeCheckError:
// main/const_typed_overflow3.gno:9:10: e * 4 (constant 18446744073709551616 of type int64) overflows int64
