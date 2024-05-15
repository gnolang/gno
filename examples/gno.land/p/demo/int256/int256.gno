// This package provides a 256-bit signed integer type, Int, and associated functions.
package int256

import (
	"gno.land/p/demo/uint256"
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
func (z *Int) SetString(s string) (*Int, error) {
	neg := false
	// Remove max one leading +
	if len(s) > 0 && s[0] == '+' {
		neg = false
		s = s[1:]
	}

	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var (
		abs *uint256.Uint
		err error
	)
	abs, err = uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}

	return &Int{
		abs,
		neg,
	}, nil
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
func (z *Int) initiateAbs() {
	if z == nil || z.abs == nil {
		z.abs = new(uint256.Uint)
	}
}
