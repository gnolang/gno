// Ported from https://github.com/holiman/uint256
// This package provides a 256-bit unsigned integer type, Uint256, and associated functions.
package uint256

import (
	"errors"
	"math/bits"
)

const (
	MaxUint64 = 1<<64 - 1
	uintSize  = 32 << (^uint(0) >> 63)
)

// Uint is represented as an array of 4 uint64, in little-endian order,
// so that Uint[3] is the most significant, and Uint[0] is the least significant
type Uint struct {
	arr [4]uint64
}

// NewUint returns a new initialized Uint.
func NewUint(val uint64) *Uint {
	z := &Uint{arr: [4]uint64{val, 0, 0, 0}}
	return z
}

// Zero returns a new Uint initialized to zero.
func Zero() *Uint {
	return NewUint(0)
}

// One returns a new Uint initialized to one.
func One() *Uint {
	return NewUint(1)
}

// SetAllOne sets all the bits of z to 1
func (z *Uint) SetAllOne() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, MaxUint64, MaxUint64
	return z
}

// Set sets z to x and returns z.
func (z *Uint) Set(x *Uint) *Uint {
	*z = *x

	return z
}

// SetOne sets z to 1
func (z *Uint) SetOne() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, 1
	return z
}

const twoPow256Sub1 = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

// SetFromDecimal sets z from the given string, interpreted as a decimal number.
// OBS! This method is _not_ strictly identical to the (*big.Uint).SetString(..., 10) method.
// Notable differences:
// - This method does not accept underscore input, e.g. "100_000",
// - This method does not accept negative zero as valid, e.g "-0",
//   - (this method does not accept any negative input as valid))
func (z *Uint) SetFromDecimal(s string) (err error) {
	// Remove max one leading +
	if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	// Remove any number of leading zeroes
	if len(s) > 0 && s[0] == '0' {
		var i int
		var c rune
		for i, c = range s {
			if c != '0' {
				break
			}
		}
		s = s[i:]
	}
	if len(s) < len(twoPow256Sub1) {
		return z.fromDecimal(s)
	}
	if len(s) == len(twoPow256Sub1) {
		if s > twoPow256Sub1 {
			return ErrBig256Range
		}
		return z.fromDecimal(s)
	}
	return ErrBig256Range
}

// FromDecimal is a convenience-constructor to create an Uint from a
// decimal (base 10) string. Numbers larger than 256 bits are not accepted.
func FromDecimal(decimal string) (*Uint, error) {
	var z Uint
	if err := z.SetFromDecimal(decimal); err != nil {
		return nil, err
	}
	return &z, nil
}

// MustFromDecimal is a convenience-constructor to create an Uint from a
// decimal (base 10) string.
// Returns a new Uint and panics if any error occurred.
func MustFromDecimal(decimal string) *Uint {
	var z Uint
	if err := z.SetFromDecimal(decimal); err != nil {
		panic(err)
	}
	return &z
}

// multipliers holds the values that are needed for fromDecimal
var multipliers = [5]*Uint{
	nil, // represents first round, no multiplication needed
	{[4]uint64{10000000000000000000, 0, 0, 0}},                                     // 10 ^ 19
	{[4]uint64{687399551400673280, 5421010862427522170, 0, 0}},                     // 10 ^ 38
	{[4]uint64{5332261958806667264, 17004971331911604867, 2938735877055718769, 0}}, // 10 ^ 57
	{[4]uint64{0, 8607968719199866880, 532749306367912313, 1593091911132452277}},   // 10 ^ 76
}

// fromDecimal is a helper function to only ever be called via SetFromDecimal
// this function takes a string and chunks it up, calling ParseUint on it up to 5 times
// these chunks are then multiplied by the proper power of 10, then added together.
func (z *Uint) fromDecimal(bs string) error {
	// first clear the input
	z.Clear()
	// the maximum value of uint64 is 18446744073709551615, which is 20 characters
	// one less means that a string of 19 9's is always within the uint64 limit
	var (
		num       uint64
		err       error
		remaining = len(bs)
	)
	if remaining == 0 {
		return errors.New("EOF")
	}
	// We proceed in steps of 19 characters (nibbles), from least significant to most significant.
	// This means that the first (up to) 19 characters do not need to be multiplied.
	// In the second iteration, our slice of 19 characters needs to be multipleied
	// by a factor of 10^19. Et cetera.
	for i, mult := range multipliers {
		if remaining <= 0 {
			return nil // Done
		} else if remaining > 19 {
			num, err = parseUint(bs[remaining-19:remaining], 10, 64)
		} else {
			// Final round
			num, err = parseUint(bs, 10, 64)
		}
		if err != nil {
			return err
		}
		// add that number to our running total
		if i == 0 {
			z.SetUint64(num)
		} else {
			base := NewUint(num)
			z.Add(z, base.Mul(base, mult))
		}
		// Chop off another 19 characters
		if remaining > 19 {
			bs = bs[0 : remaining-19]
		}
		remaining -= 19
	}
	return nil
}

// Byte sets z to the value of the byte at position n,
// with 'z' considered as a big-endian 32-byte integer
// if 'n' > 32, f is set to 0
// Example: f = '5', n=31 => 5
func (z *Uint) Byte(n *Uint) *Uint {
	// in z, z.arr[0] is the least significant
	if number, overflow := n.Uint64WithOverflow(); !overflow {
		if number < 32 {
			number := z.arr[4-1-number/8]
			offset := (n.arr[0] & 0x7) << 3 // 8*(n.d % 8)
			z.arr[0] = (number & (0xff00000000000000 >> offset)) >> (56 - offset)
			z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
			return z
		}
	}

	return z.Clear()
}

// BitLen returns the number of bits required to represent z
func (z *Uint) BitLen() int {
	switch {
	case z.arr[3] != 0:
		return 192 + bits.Len64(z.arr[3])
	case z.arr[2] != 0:
		return 128 + bits.Len64(z.arr[2])
	case z.arr[1] != 0:
		return 64 + bits.Len64(z.arr[1])
	default:
		return bits.Len64(z.arr[0])
	}
}

// ByteLen returns the number of bytes required to represent z
func (z *Uint) ByteLen() int {
	return (z.BitLen() + 7) / 8
}

// Clear sets z to 0
func (z *Uint) Clear() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, 0
	return z
}

const (
	// hextable  = "0123456789abcdef"
	bintable  = "\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\x00\x01\x02\x03\x04\x05\x06\a\b\t\xff\xff\xff\xff\xff\xff\xff\n\v\f\r\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\n\v\f\r\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
	badNibble = 0xff
)

// SetFromHex sets z from the given string, interpreted as a hexadecimal number.
// OBS! This method is _not_ strictly identical to the (*big.Int).SetString(..., 16) method.
// Notable differences:
// - This method _require_ "0x" or "0X" prefix.
// - This method does not accept zero-prefixed hex, e.g. "0x0001"
// - This method does not accept underscore input, e.g. "100_000",
// - This method does not accept negative zero as valid, e.g "-0x0",
//   - (this method does not accept any negative input as valid)
func (z *Uint) SetFromHex(hex string) error {
	return z.fromHex(hex)
}

// fromHex is the internal implementation of parsing a hex-string.
func (z *Uint) fromHex(hex string) error {
	if err := checkNumberS(hex); err != nil {
		return err
	}
	if len(hex) > 66 {
		return ErrBig256Range
	}
	z.Clear()
	end := len(hex)
	for i := 0; i < 4; i++ {
		start := end - 16
		if start < 2 {
			start = 2
		}
		for ri := start; ri < end; ri++ {
			nib := bintable[hex[ri]]
			if nib == badNibble {
				return ErrSyntax
			}
			z.arr[i] = z.arr[i] << 4
			z.arr[i] += uint64(nib)
		}
		end = start
	}
	return nil
}

// FromHex is a convenience-constructor to create an Uint from
// a hexadecimal string. The string is required to be '0x'-prefixed
// Numbers larger than 256 bits are not accepted.
func FromHex(hex string) (*Uint, error) {
	var z Uint
	if err := z.fromHex(hex); err != nil {
		return nil, err
	}
	return &z, nil
}

// MustFromHex is a convenience-constructor to create an Uint from
// a hexadecimal string.
// Returns a new Uint and panics if any error occurred.
func MustFromHex(hex string) *Uint {
	var z Uint
	if err := z.fromHex(hex); err != nil {
		panic(err)
	}
	return &z
}

// Clone creates a new Uint identical to z
func (z *Uint) Clone() *Uint {
	var x Uint
	x.arr[0] = z.arr[0]
	x.arr[1] = z.arr[1]
	x.arr[2] = z.arr[2]
	x.arr[3] = z.arr[3]

	return &x
}
