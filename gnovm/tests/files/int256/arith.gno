package main

import "math"

func main() {
	// uint256 wraps modulo 2^256.
	var max uint256 = math.MaxUint256
	var one uint256 = 1
	println(max+one, max*max, one-2, -one)
	println(max/3, max%10, (max-1)/max, max&^one)

	// int256 is two's complement.
	var imax int256 = math.MaxInt256
	var imin int256 = math.MinInt256
	var m1 int256 = -1
	println(imax+1 == imin, imin-1 == imax, -imin == imin, imin*m1 == imin)
	println(imin/m1 == imin, imin%m1, imax/m1, imin/2)

	// Division truncates toward zero, and the remainder has the sign of the
	// dividend, as in Go.
	for _, p := range [][2]int256{{7, 2}, {-7, 2}, {7, -2}, {-7, -2}} {
		println(p[0], p[1], p[0]/p[1], p[0]%p[1])
	}

	// Bitwise operations.
	var x uint256 = 0xf0f0
	println(x&0xff, x|0xf, x^0xffff, x&^0xf000, ^x == max-x)
	var y int256 = -2
	println(y&3, y|1, y^-1, y&^1, ^y)

	// Increment, decrement and compound assignment.
	x = max
	x++
	println(x)
	x--
	println(x == max)
	var z int256 = 10
	z += 5
	z -= 20
	z *= -3
	z /= 4
	z %= 3
	z <<= 100
	z >>= 99
	z |= 1
	z &= 7
	z ^= 2
	z &^= 1
	println(z)

	// Comparisons.
	println(imin < imax, imax > imin, imin <= imin, m1 < 0, max > 0, uint256(imin) > uint256(imax))
}

// Output:
// 0 1 115792089237316195423570985008687907853269984665640564039457584007913129639935 115792089237316195423570985008687907853269984665640564039457584007913129639935
// 38597363079105398474523661669562635951089994888546854679819194669304376546645 5 0 115792089237316195423570985008687907853269984665640564039457584007913129639934
// true true true true
// true 0 -57896044618658097711785492504343953926634992332820282019728792003956564819967 -28948022309329048855892746252171976963317496166410141009864396001978282409984
// 7 2 3 1
// -7 2 -3 -1
// 7 -2 -3 1
// -7 -2 3 -1
// 240 61695 3855 240 true
// 2 -1 1 -2 1
// 0
// true
// 2
// true true true true true true
