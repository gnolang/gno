package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int32 = 1<<31 - 1

func main() {
	println(c + 1)
}

// Error:
// main/int32_fold.gno:9:10-15: constant 2147483648 overflows Int32Kind
