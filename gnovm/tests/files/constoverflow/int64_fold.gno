package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int64 = 1<<63 - 1

func main() {
	println(c + 1)
}

// Error:
// main/int64_fold.gno:9:10-15: constant 9223372036854775808 overflows Int64Kind
