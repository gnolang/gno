package main

// A negative ellipsis-array key is rejected at key validation ("index must not
// be negative") before the measurement loop runs, so the loop only ever sees
// non-negative keys. Pinned alongside make23/make24 to document the full set of
// out-of-range keys handled upstream of the MaxInt64 length guard.

func main() {
	x := [...]int{-1: 1}
	_ = x
}

// Error:
// main/make25.gno:9:7-22: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/make25.gno:9:16: invalid argument: index -1 (constant of type int) must not be negative
