// Kept apart from const.gno so that tools that turn Gno into Go, which has no
// 256-bit integers, can handle 256-bit code by whole files.

package math

// 256-bit integer limit values, for the int256 and uint256 types. These have
// no counterpart in Go. MaxUint256 can be used as a uint256 constant. An
// int256 constant must fit in an int64, so MaxInt256 and MinInt256 cannot be
// used where an int256 constant is expected; convert a uint256 value instead:
//
//	var bits uint256 = math.MaxInt256
//	max := int256(bits)
const (
	MaxInt256  = 1<<255 - 1 // 57896044618658097711785492504343953926634992332820282019728792003956564819967
	MinInt256  = -1 << 255  // -57896044618658097711785492504343953926634992332820282019728792003956564819968
	MaxUint256 = 1<<256 - 1 // 115792089237316195423570985008687907853269984665640564039457584007913129639935
)
