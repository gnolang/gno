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
//     arbitrary precision arithmetic.
//
// # Division and modulus operations
//
// This package provides three different division and modulus operations:
//
//   - Div and Rem: Truncated division (T-division)
//   - Quo and Mod: Floored division (F-division)
//   - DivE and ModE: Euclidean division (E-division)
//
// Truncated division (Div, Rem) is the most common implementation in modern processors
// and programming languages. It rounds quotients towards zero and the remainder
// always has the same sign as the dividend.
//
// Floored division (Quo, Mod) always rounds quotients towards negative infinity.
// This ensures that the modulus is always non-negative for a positive divisor,
// which can be useful in certain algorithms.
//
// Euclidean division (DivE, ModE) ensures that the remainder is always non-negative,
// regardless of the signs of the dividend and divisor. This has several mathematical
// advantages:
//
//  1. It satisfies the unique division with remainder theorem.
//  2. It preserves division and modulus properties for negative divisors.
//  3. It allows for optimizations in divisions by powers of two.
//
// [+] Currently, ModE and Mod are shared the same implementation.
//
// ## Performance considerations:
//
//   - For most operations, the performance difference between these division types is negligible.
//   - Euclidean division may require an extra comparison and potentially an addition,
//     which could impact performance in extremely performance-critical scenarios.
//   - For divisions by powers of two, Euclidean division can be optimized to use
//     bitwise operations, potentially offering better performance.
//
// ## Usage guidelines:
//
//   - Use Div and Rem for general-purpose division that matches most common expectations.
//   - Use Quo and Mod when you need a non-negative remainder for positive divisors,
//     or when implementing algorithms that assume floored division.
//   - Use DivE and ModE when you need the mathematical properties of Euclidean division,
//     or when working with algorithms that specifically require it.
//
// Note: When working with negative numbers, be aware of the differences in behavior
// between these division types, especially at the boundaries of integer ranges.
//
// ## References
//
// Daan Leijen, “Division and Modulus for Computer Scientists”:
// https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/divmodnote-letter.pdf
package int256
