package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int64 = -1 << 63

func main() {
	println(-c)
}

// Error:
// main/int64_neg.gno:9:10-12: constant 9223372036854775808 overflows Int64Kind
