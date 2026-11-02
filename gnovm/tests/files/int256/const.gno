package main

import "math"

// Constants at the boundaries of each type. A uint256 constant has the full
// range; an int256 constant must fit in an int64 (a language restriction),
// so the int256 boundaries are checked with values converted at run time.
const (
	maxU uint256 = 1<<256 - 1
	hex  uint256 = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	big          = 1 << 200 // untyped
	maxI int256  = math.MaxInt64
	minI int256  = math.MinInt64
)

func main() {
	println(maxU, hex == maxU, maxU == math.MaxUint256, maxI, minI)

	// Constant folding, with exact results that fit the type.
	const a = maxU / 2
	const b = minI + maxI
	const c = maxU >> 255
	const d = -maxI
	const f = ^uint256(0)
	const g = uint256(1) << 255
	println(a, b, c, d, f == maxU, g)

	// Untyped constants and float constants that are whole numbers convert
	// to uint256.
	var u uint256 = big
	var v uint256 = 1e70
	println(u, v)

	// The int256 workaround: convert a uint256 value.
	var maxBits uint256 = math.MaxInt256
	imax, imin := int256(maxBits), int256(maxBits+1)
	w := -int256(u)
	println(imax, imin, w, int256(u)*4)

	// Typed constants of other kinds convert to 256 bits exactly.
	const m64 int64 = math.MinInt64
	println(int256(m64), uint256(uint64(math.MaxUint64)))
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935 true true 9223372036854775807 -9223372036854775808
// 57896044618658097711785492504343953926634992332820282019728792003956564819967 -1 1 -9223372036854775807 true 57896044618658097711785492504343953926634992332820282019728792003956564819968
// 1606938044258990275541962092341162602522202993782792835301376 10000000000000000000000000000000000000000000000000000000000000000000000
// 57896044618658097711785492504343953926634992332820282019728792003956564819967 -57896044618658097711785492504343953926634992332820282019728792003956564819968 -1606938044258990275541962092341162602522202993782792835301376 6427752177035961102167848369364650410088811975131171341205504
// -9223372036854775808 18446744073709551615
