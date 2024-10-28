// The int256 package provides a 256-bit signed interger type for gno,
// supporting arithmetic operations and bitwise manipulation.
//
// It designed for applications that require high-precision arithmetic
// beyond the standard 64-bit range.
//
// ## Features
//
//   - 256-bit Signed Integers: Support for large integer ranging from -2^255 to 2^255-1.
//   - Two's Complement Representation: Efficient storage and computation using two's complement.
//   - Arithmetic Operations: Add, Sub, Mul, Div, Mod, Inc, Dec, etc.
//   - Bitwise Operations: And, Or, Xor, Not, etc.
//   - Comparison Operations: Cmp, Eq, Lt, Gt, etc.
//   - Conversion Functions: Int to Uint, Uint to Int, etc.
//   - String Parsing and Formatting: Convert to and from decimal string representation.
//
// ## Notes
//
//   - Some methods may panic when encountering invalid inputs or overflows.
//   - The `int256.Int` type can interact with `uint256.Uint` from the `p/demo/uint256` package.
//   - Unlike `math/big.Int`, the `int256.Int` type has fixed size (256-bit) and does not support
//
// arbitrary precision arithmetic.
package int256
