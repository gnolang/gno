package main

import "math/big"

// Exercises math/big from user code (end-to-end import path) and covers
// SetUint64/Uint64, which have no direct in-package test.
func main() {
	// SetUint64 must stay positive even above MaxInt64.
	z := new(big.Int).SetUint64(1 << 63)
	println(z.String(), z.Sign())

	// MaxUint64 round-trips through Uint64.
	const maxU64 = uint64(18446744073709551615)
	r := new(big.Int).SetUint64(maxU64).Uint64()
	println(r == maxU64)

	// A heavy op reached through the native bridge.
	a, _ := new(big.Int).SetString("100000000000000000000", 10)
	a.Mul(a, big.NewInt(3))
	println(a.String())
}

// Output:
// 9223372036854775808 1
// true
// 300000000000000000000
