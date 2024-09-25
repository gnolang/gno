package int256

import (
	"gno.land/p/demo/uint256"
)

var one = uint256.NewUint(1)

type Int struct {
	value uint256.Uint
}

// New creates and returns a new Int initialized to zero.
func New() *Int {
	return &Int{}
}

// NewInt allocates and returns a new Int set to the value of the provided int64.
func NewInt(x int64) *Int {
	return new(Int).SetInt64(x)
}

// Zero returns a new Int initialized to 0.
//
// This function is useful for creating a starting point for calculations or
// when an explicit zero value is needed.
func Zero() *Int {
	return &Int{}
}

// One returns a new Int initialized to one.
//
// This function is convenient for operations that require a unit value,
// such as incrementing or serving as an identity element in multiplication.
func One() *Int {
	z := &Int{}
	z.value.SetUint64(1)
	return z
}

// Sign determines the sign of the Int.
//
// It returns -1 for negative numbers, 0 for zero, and +1 for positive numbers.
func (z *Int) Sign() int {
	if z.value.IsZero() {
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
	temp.Rsh(&z.value, 255)
	if temp.IsZero() {
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
	return new(Int).SetString(s)
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
//
// For negative numbers, it uses the two's complement representation:
// 1. Convert the absolute value to binary
// 2. Invert all bits (NOT operation)
// 3. Add 1 to the result
//
// This process ensures correct representation and arithmetic for negative numbers.
func (z *Int) SetString(s string) (*Int, error) {
	neg := false
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		if s[0] == '-' {
			neg = true
		}
		// remove the sign character from the string
		// and take only the numeric part
		s = s[1:]
	}

	temp, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}

	// for negative numbers, apply two's complement
	if neg {
		// invert all bits (NOT operation)
		temp.Not(temp)
		// add one (two's complement)
		temp.Add(temp, uint256.NewUint(1))
		// Example of two's complement for 8-bit number -5:
		//
		// Decimal: 5
		// Step 0 (binary):        0000 0101
		// Step 1 (NOT):           1111 1010
		// Step 2 (Add 1):         1111 1011
		//
		// Result: 1111 1011 (binary representation of -5 in two's complement)
		//
		// This process scales to our 256-bit integers, allowing
		// for correct representation and arithmetic of negative numbers.
	}

	z.value.Set(temp)
	return z, nil
}

// SetInt64 sets the Int to the value of the provided int64.
//
// This method allows for easy conversion from standard Go integer types
// to Int, correctly handling both positive and negative values.
func (z *Int) SetInt64(v int64) *Int {
	if v >= 0 {
		z.value.SetUint64(uint64(v))
	} else {
		z.value.SetUint64(uint64(-v))
		z.value.Not(&z.value)
		z.value.Add(&z.value, one)
	}
	return z
}

// SetUint64 sets the Int to the value of the provided uint64.
func (z *Int) SetUint64(v uint64) *Int {
	z.value.SetUint64(v)
	return z
}

// FromUint256 sets the Int to the value of the provided Uint256.
//
// This method allows for conversion from unsigned 256-bit integers
// to signed integers.
func (z *Int) FromUint256(v *uint256.Uint) *Int {
	z.value.Set(v)
	return z
}

// NilToZero returns the Int if it's not nil, or a new zero-valued Int otherwise.
//
// This method is useful for safely handling potentially nil Int pointers,
// ensuring that operations always have a valid Int to work with.
func (z *Int) NilToZero() *Int {
	if z == nil {
		return Zero()
	}
	return z
}
