package main

// A key one past MaxInt64 cannot be represented as an int index at all, so it
// is rejected at key conversion ("bigint overflows target kind") before the
// ellipsis-array measurement runs. This is why the MaxInt64 guard in make23
// only needs to handle k == MaxInt64: any larger key never reaches it.

func main() {
	x := [...]int{9223372036854775808: 1}
	_ = x
}

// Error:
// main/make24.gno:9:7-39: bigint overflows target kind

// TypeCheckError:
// main/make24.gno:9:16: 9223372036854775808 (untyped int constant) overflows int
