package main

// A typed uint64 constant sum past the type is rejected, not zeroed
// (#6392).

const f uint64 = 1 << 63

func main() {
	println(f + f)
}

// Error:
// main/const_typed_overflow4.gno:9:10-15: f + f (constant 18446744073709551616 of type uint64) overflows uint64

// TypeCheckError:
// main/const_typed_overflow4.gno:9:10: f + f (constant 18446744073709551616 of type uint64) overflows uint64
