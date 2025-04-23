package uint256

import (
	"encoding/binary"
	"errors"
	"math/bits"
	"strconv"
	"strings"
)

type Uint [4]uint64

// usefull

func reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func lpad(s string, totalLength int, padChar rune) string {
	padCount := totalLength - len(s)
	if padCount > 0 {
		return strings.Repeat(string(padChar), padCount) + s
	}
	return s
}

// ctors

func NewFromString(s string, base int) (*Uint, error) {
	u := New()
	for _, c := range s {
		var digit uint64
		switch {
		case '0' <= c && c <= '9':
			digit = uint64(c - '0')
		case 'a' <= c && c <= 'z':
			digit = uint64(c - 'a' + 10)
		case 'A' <= c && c <= 'Z':
			digit = uint64(c - 'A' + 10)
		default:
			return nil, errors.New("invalid digit: " + string(c))
		}
		if digit >= uint64(base) {
			return nil, errors.New("digit out of range for base " + strconv.Itoa(base) + ": " + string(c))
		}
		u.Mul64(uint64(base))
		u.Add64(digit)
		// fmt.Printf(">%#x, %#x, %#x, %#x\n", u[0], u[1], u[2], u[3])
	}
	return u, nil
}

func New() *Uint {
	return (&Uint{}).Clear()
}

func NewFrom64s(us ...uint64) *Uint {
	u := New()
	for n := 0; n < len(us) && n < 4; n++ {
		u[n] = us[n]
	}
	return u
}

func NewFromBytes(b []byte) *Uint {
	u := New()
	if len(b) > 32 {
		return u
	}
	lenBlocs := len(b) / 8
	if len(b)%8 != 0 {
		lenBlocs += 1
	}
	blocs := [4][]byte{}
	for i := 0; i < lenBlocs; i++ {
		l := len(b)
		if l < 8 {
			blocs[i] = append(make([]byte, 8-l), b...)
			b = []byte{}
		} else {
			blocs[i] = b[l-8:]
			b = b[:l-8]
		}
	}
	for i := 0; i < lenBlocs; i++ {
		u[i] = binary.BigEndian.Uint64(blocs[i])
	}
	return u
}
func (u Uint) HexString() string {
	if u.IsZero() {
		return "0x0"
	}

	lastNonZero := 3
	for i := 3; i >= 0; i-- {
		if u[i] != 0 {
			break
		}
		lastNonZero = i
	}

	result := "0x"
	for i := lastNonZero; i >= 0; i-- {
		hex := strconv.FormatUint(u[i], 16)
		if i < lastNonZero {
			for len(hex) < 16 {
				hex = "0" + hex
			}
		}
		result += hex
	}

	return result
}

func (u Uint) divMod10() (Uint, uint64) {
	var result Uint
	var remainder uint64 = 0

	for i := 3; i >= 0; i-- {
		high := remainder
		low := u[i]
		quotient, rem := div128by10(high, low)
		result[i] = quotient
		remainder = rem
	}
	return result, remainder
}

func div128by10(hi, lo uint64) (quotient, remainder uint64) {
	const divisor = 10
	var n uint64 = 64
	var q uint64 = 0
	for n > 0 {
		n--
		bit := (lo >> 63) & 1
		hi = (hi << 1) | bit
		lo <<= 1
		q <<= 1
		if hi >= divisor {
			hi -= divisor
			q |= 1
		}
	}
	return q, hi
}

func (u Uint) String() string {
	if u.IsZero() {
		return "0"
	}

	var digits []byte
	n := u

	for !n.IsZero() {
		var rem uint64
		n, rem = n.divMod10()
		digits = append(digits, '0'+byte(rem))
	}

	// Reverse the digits
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}

// arithmetics

func carryAdd(a uint64, b uint64) (uint64, uint64) {
	var carry uint64
	sum := a + b
	if sum < a || sum < b {
		carry = 1
	}
	return sum, carry
}

func carryMul(a, b uint64) (lo, hi uint64) {
	const mask32 = 0xFFFFFFFF
	a0 := a & mask32
	a1 := a >> 32
	b0 := b & mask32
	b1 := b >> 32
	p0 := a0 * b0
	p1 := a0 * b1
	p2 := a1 * b0
	p3 := a1 * b1
	mid1 := p1 + (p0 >> 32)
	carry1 := uint64(0)
	if mid1 < p1 {
		carry1++
	}
	mid2 := p2 + (mid1 & mask32)
	carry2 := uint64(0)
	if mid2 < p2 {
		carry2++
	}
	lo = (mid2 << 32) | (p0 & mask32)
	hi = p3 + (mid1 >> 32) + (mid2 >> 32) + carry1 + carry2
	return lo, hi
}

func (u *Uint) add64Level(x uint64, n int) {
	for i := n; true; i++ {
		res, carry := carryAdd(u[i], x)
		u[i] = res
		if carry == 0 {
			break
		}
		x = carry
	}
}

func (u *Uint) Add64(x uint64) {
	u.add64Level(x, 0)
}

func (x *Uint) Add256(y *Uint) *Uint {
	var carry uint64
	x[0], carry = bits.Add64(x[0], y[0], 0)
	x[1], carry = bits.Add64(x[1], y[1], carry)
	x[2], carry = bits.Add64(x[2], y[2], carry)
	x[3], _ = bits.Add64(x[3], y[3], carry)
	return x
}
func (x *Uint) Mul256(y *Uint) *Uint {
	result := New()
	for i := 0; i < 4; i++ {
		if x[i] == 0 {
			continue
		}
		for j := 0; j < 4; j++ {
			if i+j >= 4 {
				break // skip overflow >256 bits
			}
			lo, hi := carryMul(x[i], y[j])
			result.add64Level(lo, i+j)
			if i+j+1 < 4 {
				result.add64Level(hi, i+j+1)
			}
		}
	}
	x.AssignValue(*result)
	return x
}

func (x *Uint) Sub256(y Uint) *Uint {
	var carry uint64
	x[0], carry = bits.Sub64(x[0], y[0], 0)
	x[1], carry = bits.Sub64(x[1], y[1], carry)
	x[2], carry = bits.Sub64(x[2], y[2], carry)
	x[3], _ = bits.Sub64(x[3], y[3], carry)
	return x
}

func (x *Uint) Div256(y *Uint) *Uint {
	if y.IsZero() || y.Greater(*x) {
		return x.Clear()
	}
	if x.Equals(y) {
		return x.AssignValue(Uint{1})
	}
	if x.IsUint64() {
		return x.SetUint64(x[3] / y[3])
	}

	var quotient, remainder Uint
	dividend := *x
	divisor := *y

	for i := 255; i >= 0; i-- {
		remainder = remainder.lsh(1)

		if dividend.GetBit(i) {
			remainder[0] |= 1
		}

		if !divisor.Greater(remainder) {
			remainder.Sub256(divisor)
			quotient.SetBit(i)
		}
	}

	*x = quotient
	return x
}

func (u *Uint) Mul64(x uint64) {
	dup := New()
	for i := 0; i < 3; i++ {
		lo, hi := carryMul(u[i], x)
		dup.add64Level(lo, i)
		if hi == 0 {
			continue
		}
		if i == 3 { // cant add more than 256 bits
			break
		}
		dup.add64Level(hi, i+1)
	}
	u.AssignValue(*dup)
}

func (u Uint) Greater(x Uint) bool {
	return !x.Lesser(u)
}

func (z Uint) Lesser(x Uint) bool {
	_, carry := bits.Sub64(z[0], x[0], 0)
	_, carry = bits.Sub64(z[1], x[1], carry)
	_, carry = bits.Sub64(z[2], x[2], carry)
	_, carry = bits.Sub64(z[3], x[3], carry)
	return carry != 0
}

// bit gets the i-th bit
func (x Uint) GetBit(i int) bool {
	word := i / 64
	bit := uint(i % 64)
	return (x[word] & (1 << bit)) != 0
}

// setBit sets the i-th bit to 1
func (x *Uint) SetBit(i int) {
	word := i / 64
	bit := uint(i % 64)
	x[word] |= 1 << bit
}

// lsh performs left shift by 1
func (x Uint) lsh(n uint) (result Uint) {
	for i := 3; i >= 0; i-- {
		if i < 3 {
			result[i+1] |= x[i] >> 63
		}
		result[i] = x[i] << 1
	}
	return
}

func (u Uint) IsZero() bool {
	return u[0] == 0 && u[1] == 0 && u[2] == 0 && u[3] == 0
}

func (u *Uint) IsUint64() bool {
	return (u[1] | u[2] | u[3]) == 0
}

func (u *Uint) AssignValue(x Uint) *Uint {
	u[0], u[1], u[2], u[3] = x[0], x[1], x[2], x[3]
	return u
}

func (u *Uint) SetUint64(x uint64) *Uint {
	u[0], u[1], u[2], u[3] = 0, 0, 0, x
	return u
}

func (u *Uint) Clear() *Uint {
	u[0], u[1], u[2], u[3] = 0, 0, 0, 0
	return u
}

func (x Uint) Equals(y *Uint) bool {
	return x[0] == y[0] && x[1] == y[1] && x[2] == y[2] && x[3] == y[3]
}

func (x Uint) To64() uint64 {
	return x[0]
}

func (x Uint) ToBytes() [32]byte {
	var result [32]byte
	binary.BigEndian.PutUint64(result[0:8], x[3])
	binary.BigEndian.PutUint64(result[8:16], x[2])
	binary.BigEndian.PutUint64(result[16:24], x[1])
	binary.BigEndian.PutUint64(result[24:32], x[0])
	return result
}
