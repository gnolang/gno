package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int32 = -1 << 31

func main() {
	println(-c)
}

// Error:
// main/int32_neg.gno:9:10-12: constant 2147483648 overflows Int32Kind
