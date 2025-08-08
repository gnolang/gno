package int256

import (
	"errors"

	"gno.land/p/onbloc/uint256"
)

var (
	int1  = NewInt(1)
	uint0 = uint256.NewUint(0)
	uint1 = uint256.NewUint(1)
)

type Int struct {
	value uint256.Uint
}

// New creates and returns a new Int initialized to zero.
func New() *Int {
	return &Int{}
}

// NewInt allocates and returns a new Int set to the value of the provided int64.
func NewInt(x int64) *Int {
	return New().SetInt64(x)
}

// Zero returns a new Int initialized to 0.
//
// This function is useful for creating a starting point for calculations or
// when an explicit zero value is needed.
func Zero() *Int { return &Int{} }

// One returns a new Int initialized to one.
//
// This function is convenient for operations that require a unit value,
// such as incrementing or serving as an identity element in multiplication.
func One() *Int {
	return &Int{
		value: *uint256.NewUint(1),
	}
}

// Sign determines the sign of the Int.
//
// It returns -1 for negative numbers, 0 for zero, and +1 for positive numbers.
func (z *Int) Sign() int {
	if z == nil || z.IsZero() {
		return 0
	}
	// Right shift the value by 255 bits to check the sign bit.
	// In two's complement representation, the most significant bit (MSB) is the sign bit.
	// If the MSB is 0, the number is positive; if it is 1, the number is negative.
	//
	// Example:
	// Original value:  1 0 1 0 ... 0 1  (256 bits)
	// After Rsh 255:   0 0 0 0 ... 0 1  (1 bit)
	//
	// This approach is highly efficient as it avoids the need for comparisons
	// or arithmetic operations on the full 256-bit number. Instead it reduces
	// the problem to checking a single bit.
	//
	// Additionally, this method will work correctly for all values,
	// including the minimum possible negative number (which in two's complement
	// doesn't have a positive counterpart in the same bit range).
	var temp uint256.Uint
	if temp.Rsh(&z.value, 255).IsZero() {
		return 1
	}
	return -1
}

// FromDecimal creates a new Int from a decimal string representation.
// It handles both positive and negative values.
//
// This function is useful for parsing user input or reading numeric data
// from text-based formats.
func FromDecimal(s string) (*Int, error) {
	return New().SetString(s)
}

// MustFromDecimal is similar to FromDecimal but panics if the input string
// is not a valid decimal representation.
func MustFromDecimal(s string) *Int {
	z, err := FromDecimal(s)
	if err != nil {
		panic(err)
	}
	return z
}

// SetString sets the Int to the value represented by the input string.
// This method supports decimal string representations of integers and handles
// both positive and negative values.
func (z *Int) SetString(s string) (*Int, error) {
	if len(s) == 0 {
		return nil, errors.New("cannot set int256 from empty string")
	}

	// Check for negative sign
	neg := s[0] == '-'
	if neg || s[0] == '+' {
		s = s[1:]
	}

	// Convert string to uint256
	temp, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}

	// If negative, negate the uint256 value
	if neg {
		temp.Neg(temp)
	}

	z.value.Set(temp)
	return z, nil
}

// FromUint256 sets the Int to the value of the provided Uint256.
//
// This method allows for conversion from unsigned 256-bit integers
// to signed integers.
func (z *Int) FromUint256(v *uint256.Uint) *Int {
	z.value.Set(v)
	return z
}
