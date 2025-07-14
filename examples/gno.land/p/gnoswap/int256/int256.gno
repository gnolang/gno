// This package provides a 256-bit signed integer type, Int, and associated functions.
package int256

import (
	"errors"

	"gno.land/p/gnoswap/uint256"
)

var (
	errEmptyString                 = errors.New("empty string")
	errStringContainsOnlySign      = errors.New("string contains only sign")
	errInvalidSignInMiddleOfNumber = errors.New("invalid sign in middle of number")
)

var one = uint256.NewUint(1)

type Int struct {
	abs *uint256.Uint
	neg bool
}

// Zero returns a new Int set to 0.
func Zero() *Int {
	return NewInt(0)
}

// One returns a new Int set to 1.
func One() *Int {
	return NewInt(1)
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (z *Int) Sign() int {
	z.initiateAbs()

	if z.abs.IsZero() {
		return 0
	}
	if z.neg {
		return -1
	}
	return 1
}

// New returns a new Int set to 0.
func New() *Int {
	return &Int{
		abs: new(uint256.Uint),
	}
}

// NewInt allocates and returns a new Int set to x.
func NewInt(x int64) *Int {
	return New().SetInt64(x)
}

// FromDecimal returns a new Int from a decimal string.
// Returns a new Int and an error if the string is not a valid decimal.
func FromDecimal(s string) (*Int, error) {
	return new(Int).SetString(s)
}

// MustFromDecimal returns a new Int from a decimal string.
// Panics if the string is not a valid decimal.
func MustFromDecimal(s string) *Int {
	z, err := FromDecimal(s)
	if err != nil {
		panic(err)
	}
	return z
}

// SetString sets s to the value of z and returns z and a boolean indicating success.
//
// Note: Since we only need to check for sign characters ('+', '-'), we use a bit masking technique
// instead of [strings.IndexAny]. Although the environment is not exactly the same as gno,
// this approach was adopted because it showed overwhelmingly better performance compared to
// other implementations when benchmarked in go (with gcflags set to "-N -l").
//
// Reference: "Bit Twiddling Hacks" by Sean Eron Anderson
// (http://graphics.stanford.edu/~seander/bithacks.html)
// "Determine if a word has a byte equal to n" technique is applied.
func (z *Int) SetString(s string) (*Int, error) {
	if len(s) == 0 {
		return nil, errEmptyString
	}

	// check sign only in the first character
	neg := false
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		neg = true
		s = s[1:]
	}

	// check if the string is empty after removing the sign
	if len(s) == 0 {
		return nil, errStringContainsOnlySign
	}

	// Parallel comparison technique for sign detection
	// Process in 8-byte chunks
	sLen := len(s)
	i := 0

	// Process 8 bytes at a time
	for i+7 < sLen {
		// Convert 8 bytes into a single uint64
		// This method processes bytes directly, so no endianness issues
		//
		// access up to s[i+7] is safe, then we can reduce the number of bounds checks
		_ = s[i+7]
		chunk := uint64(s[i]) | uint64(s[i+1])<<8
		chunk |= uint64(s[i+2])<<16 | uint64(s[i+3])<<24
		chunk |= uint64(s[i+4])<<32 | uint64(s[i+5])<<40
		chunk |= uint64(s[i+6])<<48 | uint64(s[i+7])<<56

		// Operation to check for '+' (0x2B)
		// Subtracting 0x2B from each byte makes '+' bytes become 0
		// Subtracting 0x01 makes bytes in ASCII range (0-127) have 0 in their highest bit
		// Therefore, AND with 0x80 to check for zero bytes
		plusTest := ((chunk ^ 0x2B2B2B2B2B2B2B2B) - 0x0101010101010101) & 0x8080808080808080

		// check for '-' (0x2D)
		minusTest := ((chunk ^ 0x2D2D2D2D2D2D2D2D) - 0x0101010101010101) & 0x8080808080808080

		// If either test is non-zero, a sign character exists
		if (plusTest | minusTest) != 0 {
			return nil, errInvalidSignInMiddleOfNumber
		}

		i += 8
	}

	// Process remaining bytes
	for ; i < sLen; i++ {
		if s[i] == '+' || s[i] == '-' {
			return nil, errInvalidSignInMiddleOfNumber
		}
	}

	abs, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}

	return &Int{abs, neg}, nil
}

// FromUint256 is a convenience-constructor from uint256.Uint.
// Returns a new Int and whether overflow occurred.
// OBS: If u is `nil`, this method returns `nil, false`
func FromUint256(x *uint256.Uint) *Int {
	if x == nil {
		return nil
	}
	z := Zero()

	z.SetUint256(x)
	return z
}

// OBS, differs from original mempooler int256
// NilToZero sets z to 0 and return it if it's nil, otherwise it returns z
func (z *Int) NilToZero() *Int {
	if z == nil {
		return NewInt(0)
	}
	return z
}

// initiateAbs sets default value for `z` or `z.abs` value if is nil
// OBS: differs from mempooler int256. It checks not only `z.abs` but also `z`
func (z *Int) initiateAbs() *Int {
	if z == nil {
		return NewInt(0)
	}
	if z.abs == nil {
		z.abs = new(uint256.Uint)
	}
	return z
}
