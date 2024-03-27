// ported from github.com/mempooler/int256
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
	if len(z.abs) == 0 {
		return 0
	}
	if z.neg {
		return -1
	}
	return 1
}

func New() *Int {
	return &Int{
		abs: new(uint256.Uint),
	}
}

// SetInt64 sets z to x and returns z.
func (z *Int) SetInt64(x int64) *Int {
	neg := false
	if x < 0 {
		neg = true
		x = -x
	}
	if z.abs == nil {
		panic("abs is nil")
	}
	z.abs = z.abs.SetUint64(uint64(x))
	z.neg = neg
	return z
}

// SetUint64 sets z to x and returns z.
func (z *Int) SetUint64(x uint64) *Int {
	if z.abs == nil {
		panic("abs is nil")
	}
	z.abs = z.abs.SetUint64(x)
	z.neg = false
	return z
}

// NewInt allocates and returns a new Int set to x.
func NewInt(x int64) *Int {
	return New().SetInt64(x)
}

func FromDecimal(s string) (*Int, error) {
	return new(Int).SetString(s)
}

func MustFromDecimal(s string) *Int {
	z, err := FromDecimal(s)
	if err != nil {
		panic(err)
	}
	return z
}

// SetString sets s to the value of z and returns z and a boolean indicating success.
func (z *Int) SetString(s string) (*Int, error) {
	origin := s
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

// // setFromScanner implements SetString given an io.ByteScanner.
// // For documentation see comments of SetString.
// func (z *Int) setFromScanner(r io.ByteScanner, base int) (*Int, bool) {
// 	if _, _, err := z.scan(r, base); err != nil {
// 		return nil, false
// 	}
// 	// entire content must have been consumed
// 	if _, err := r.ReadByte(); err != io.EOF {
// 		return nil, false
// 	}
// 	return z, true // err == io.EOF => scan consumed all content of r
// }

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

// initiateAbs sets default value for `z.abs` value if is nil
func (z *Int) initiateAbs() {
	if z.abs == nil {
		z.abs = new(uint256.Uint)
	}
}
