// Package overflow offers overflow-checked integer arithmetic operations
// for all signed and unsigned integer types. Each of the operations returns a
// result,bool combination. The modulo % operation is not present, as it is
// always safe.
//
// This implementation is based on its Go equivalent in tm2/pkg/overflow, using
// code generation in place of generics. All functions are named matching the
// following regex:
//
//	(Add|Sub|Mul|Div)(u)?(8|16|32|64|256)?(p)?
//
// The "u" identifies a function which works on an unsigned value. The number
// identifies the bit size, and can be omitted for the int and uint types;
// 256 is for the int256 and uint256 types.
// The trailing "p" identifies the panicking versions of each function.
//
// The boolean is true when the operation did not overflow. When it is false,
// Add, Sub and Mul return the wrapped result, Div returns 0 for a zero
// divisor and the wrapped quotient (the smallest value) for the smallest
// value divided by -1. The "p" versions panic instead of returning false.
package overflow

// minInt256 is the smallest int256. An int256 constant must fit an int64, so
// it is a variable, converted from uint256 bits.
var minInt256 = int256(minInt256Bits)

var minInt256Bits uint256 = 1 << 255
