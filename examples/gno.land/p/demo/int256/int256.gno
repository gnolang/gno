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

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (z *Int) Cmp(x *Int) (r int) {
	// x cmp y == x cmp y
	// x cmp (-y) == x
	// (-x) cmp y == y
	// (-x) cmp (-y) == -(x cmp y)
	switch {
	case z == x:
		// nothing to do
	case z.neg == x.neg:
		r = z.abs.Cmp(x.abs)
		if z.neg {
			r = -r
		}
	case z.neg:
		r = -1
	default:
		r = 1
	}
	return
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

// Abs returns |z|
func (z *Int) Abs() *uint256.Uint {
	return z.abs.Clone()
}

// AbsGt returns true if |z| > x, where x is a uint256
func (z *Int) AbsGt(x *uint256.Uint) bool {
	return z.abs.Gt(x)
}

// AbsLt returns true if |z| < x, where x is a uint256
func (z *Int) AbsLt(x *uint256.Uint) bool {
	return z.abs.Lt(x)
}

// AddUint256 set z to the sum x + y, where y is a uint256, and returns z
func (z *Int) AddUint256(x *Int, y *uint256.Uint) *Int {
	if x.neg {
		if x.abs.Gt(y) {
			z.abs.Sub(x.abs, y)
			z.neg = true
		} else {
			z.abs.Sub(y, x.abs)
			z.neg = false
		}
	} else {
		z.abs.Add(x.abs, y)
		z.neg = false
	}
	return z
}

// Clone creates a new Int identical to z
func (z *Int) Clone() *Int {
	return &Int{z.abs.Clone(), z.neg}
}

// DivUint256 sets z to the quotient x/y, where y is a uint256, and returns z
// If y == 0, z is set to 0
func (z *Int) DivUint256(x *Int, y *uint256.Uint) *Int {
	z.abs.Div(x.abs, y)
	if z.abs.IsZero() {
		z.neg = false
	} else {
		z.neg = x.neg
	}
	return z
}

// Eq returns true if z == x
func (z *Int) Eq(x *Int) bool {
	return (z.neg == x.neg) && z.abs.Eq(x.abs)
}

func (z *Int) Neq(x *Int) bool {
	return !z.Eq(x)
}

// IsZero returns true if z == 0
func (z *Int) IsZero() bool {
	return z.abs.IsZero()
}

// IsNeg returns true if z < 0
func (z *Int) IsNeg() bool {
	return z.neg
}

// Lt returns true if z < x
func (z *Int) Lt(x *Int) bool {
	if z.neg {
		if x.neg {
			return z.abs.Gt(x.abs)
		} else {
			return true
		}
	} else {
		if x.neg {
			return false
		} else {
			return z.abs.Lt(x.abs)
		}
	}
}

// Gt returns true if z > x
func (z *Int) Gt(x *Int) bool {
	if z.neg {
		if x.neg {
			return z.abs.Lt(x.abs)
		} else {
			return false
		}
	} else {
		if x.neg {
			return true
		} else {
			return z.abs.Gt(x.abs)
		}
	}
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, z is set to 0 (OBS: differs from the big.Int)
func (z *Int) Mod(x, y *Int) *Int {
	if x.neg {
		z.abs.Div(x.abs, y.abs)
		z.abs.Add(z.abs, one)
		z.abs.Mul(z.abs, y.abs)
		z.abs.Sub(z.abs, x.abs)
		z.abs.Mod(z.abs, y.abs)
	} else {
		z.abs.Mod(x.abs, y.abs)
	}
	z.neg = false
	return z
}

// MulUint256 sets z to the product x*y, where y is a uint256, and returns z
func (z *Int) MulUint256(x *Int, y *uint256.Uint) *Int {
	z.abs.Mul(x.abs, y)
	if z.abs.IsZero() {
		z.neg = false
	} else {
		z.neg = x.neg
	}
	return z
}

// Neg sets z to -x and returns z.)
func (z *Int) Neg(x *Int) *Int {
	z.abs.Set(x.abs)
	if z.abs.IsZero() {
		z.neg = false
	} else {
		z.neg = !x.neg
	}
	return z
}

// Set sets z to x and returns z.
func (z *Int) Set(x *Int) *Int {
	z.abs.Set(x.abs)
	z.neg = x.neg
	return z
}

// SetFromUint256 converts a uint256.Uint to Int and sets the value to z.
func (z *Int) SetUint256(x *uint256.Uint) *Int {
	z.abs.Set(x)
	z.neg = false
	return z
}

// SubUint256 set z to the difference x - y, where y is a uint256, and returns z
func (z *Int) SubUint256(x *Int, y *uint256.Uint) *Int {
	if x.neg {
		z.abs.Add(x.abs, y)
		z.neg = true
	} else {
		if x.abs.Lt(y) {
			z.abs.Sub(y, x.abs)
			z.neg = true
		} else {
			z.abs.Sub(x.abs, y)
			z.neg = false
		}
	}
	return z
}

// Uint64 returns the lower 64-bits of z
func (z *Int) Uint64() uint64 {
	return z.abs.Uint64()
}

// Int64 returns the lower 64-bits of z
func (z *Int) Int64() int64 {

	_abs := z.abs.Clone()

	if z.neg {
		return -int64(_abs.Uint64())
	}
	return int64(_abs.Uint64())
}

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
func AddDelta(z, x *uint256.Uint, y *Int) {
	if y.neg {
		z.Sub(x, y.abs)
	} else {
		z.Add(x, y.abs)
	}
}

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
func AddDeltaOverflow(z, x *uint256.Uint, y *Int) bool {
	var overflow bool
	if y.neg {
		_, overflow = z.SubOverflow(x, y.abs)
	} else {
		_, overflow = z.AddOverflow(x, y.abs)
	}
	return overflow
}

// OBS, differs from original mempooler int256
// ToString returns the decimal representation of z.
func (z *Int) ToString() string {
	if z == nil {
		panic("int256: nil pointer to ToString()")
	}

	t := z.abs.Dec()
	if z.neg {
		return "-" + t
	}
	return t
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
