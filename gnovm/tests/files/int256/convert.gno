package main

import "math"

func main() {
	// Into the 256-bit kinds: sign extension from signed kinds, zero
	// extension from unsigned ones.
	var i8 int8 = -1
	var i16 int16 = -2
	var i32 int32 = -3
	var i64 int64 = math.MinInt64
	var i int = -5
	var u8 uint8 = 255
	var u16 uint16 = 65535
	var u32 uint32 = math.MaxUint32
	var u64 uint64 = math.MaxUint64
	var u uint = 7
	println(int256(i8), int256(i16), int256(i32), int256(i64), int256(i))
	println(uint256(i8) == math.MaxUint256, uint256(i64))
	println(int256(u8), int256(u16), int256(u32), int256(u64), int256(u))
	println(uint256(u8), uint256(u64))

	// Out of the 256-bit kinds: truncation to the low bits.
	var big uint256 = 1<<200 + 1<<64 + 0x1ff
	var neg int256 = -(1 << 70) - 2
	println(int8(big), int16(big), int32(big), int64(big), int(big))
	println(uint8(big), uint16(big), uint32(big), uint64(big), uint(big))
	println(int8(neg), int16(neg), int32(neg), int64(neg), int(neg))
	println(uint8(neg), uint16(neg), uint32(neg), uint64(neg), uint(neg))

	// Between int256 and uint256: same bits.
	var allOnes uint256 = math.MaxUint256
	var half uint256 = 1 << 255
	var m1 int256 = -1
	var imin int256 = math.MinInt256
	println(int256(allOnes), uint256(m1) == math.MaxUint256)
	println(int256(half), uint256(imin))

	// Floats: rounding to nearest for integer to float, truncation toward
	// zero for float to integer.
	println(float64(big), float32(big), float64(neg), float32(neg))
	println(float64(allOnes), float32(allOnes)) // float32 overflows to +Inf
	var f float64 = -12345.75
	var g float32 = 1e30
	println(int256(f), uint256(g), int256(-g))
	var huge float64 = 1e300
	println(int256(huge), uint256(huge), uint256(f))

	// Integer to string follows Go's rune rules.
	println(string(int256(0x263a)), string(uint256(65)), string(int256(-1)) == "�", string(uint256(1<<100)) == "�")
}

// Output:
// -1 -2 -3 -9223372036854775808 -5
// true 115792089237316195423570985008687907853269984665640564039448360635876274864128
// 255 65535 4294967295 18446744073709551615 7
// 255 18446744073709551615
// -1 511 511 511 511
// 255 511 511 511 511
// -2 -2 -2 -2 -2
// 254 65534 4294967294 18446744073709551614 18446744073709551614
// -1 true
// -57896044618658097711785492504343953926634992332820282019728792003956564819968 57896044618658097711785492504343953926634992332820282019728792003956564819968
// 1.6069380442589903e+60 +Inf -1.1805916207174113e+21 -1.1805916e+21
// 1.157920892373162e+77 +Inf
// -12345 1000000015047466219876688855040 -1000000015047466219876688855040
// 0 0 115792089237316195423570985008687907853269984665640564039457584007913129627591
// ☺ A true true
