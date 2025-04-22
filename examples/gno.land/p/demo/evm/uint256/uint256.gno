package uint256

import (
	"encoding/binary"
	"errors"
	"math/bits"
	"strconv"
	"strings"
)

type Uint struct {
	data [4]uint64
}

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
		// fmt.Printf(">%#x, %#x, %#x, %#x\n", u.data[0], u.data[1], u.data[2], u.data[3])
	}
	return u, nil
}

func New() *Uint {
	return &Uint{data: [4]uint64{0, 0, 0, 0}}
}

func NewFrom64s(us ...uint64) *Uint {
	u := New()
	for n := 0; n < len(us) && n < 4; n++ {
		u.data[n] = us[n]
	}
	return u
}

func NewFromBytes(b []byte) *Uint {
	u := New()
	lenBlocs := (len(b) / 8) + 1
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
		u.data[i] = binary.BigEndian.Uint64(blocs[i])
	}
	return u
}

// string

func (u Uint) HexString() string {
	buf := "0x"
	last0 := 3
	for i := 3; i >= 0; i-- {
		if u.data[i] == 0 {
			last0 = i
		} else {
			break
		}
	}
	if last0 == 0 {
		return "0x0"
	}
	for i := last0 - 1; i >= 0; i-- {
		n := strconv.FormatUint(u.data[i], 16)
		if i == last0-1 {
			buf += n
		} else {
			buf += lpad(n, 16, '0')
		}
	}
	return buf
}

func (u Uint) divMod10() (Uint, uint64) {
	var result Uint
	var remainder uint64 = 0

	for i := 3; i >= 0; i-- {
		high := remainder
		low := u.data[i]
		quotient, rem := div128by10(high, low)
		result.data[i] = quotient
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
		res, carry := carryAdd(u.data[i], x)
		u.data[i] = res
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
	x.data[0], carry = bits.Add64(x.data[0], y.data[0], 0)
	x.data[1], carry = bits.Add64(x.data[1], y.data[1], carry)
	x.data[2], carry = bits.Add64(x.data[2], y.data[2], carry)
	x.data[3], _ = bits.Add64(x.data[3], y.data[3], carry)
	return x
}

func (x *Uint) Mul256(y *Uint) *Uint {
	result := New()
	for i := 0; i < 4; i++ {
		if x.data[i] == 0 {
			continue
		}
		for j := 0; j < 4; j++ {
			if i+j >= 4 {
				break // skip overflow >256 bits
			}
			lo, hi := carryMul(x.data[i], y.data[j])
			result.add64Level(lo, i+j)
			if i+j+1 < 4 {
				result.add64Level(hi, i+j+1)
			}
		}
	}
	x.data = result.data
	return x
}

func (u *Uint) Mul64(x uint64) {
	dup := New()
	for i := 0; i < 3; i++ {
		lo, hi := carryMul(u.data[i], x)
		dup.add64Level(lo, i)
		if hi == 0 {
			continue
		}
		if i == 3 { // cant add more than 256 bits
			break
		}
		dup.add64Level(hi, i + 1)
	}
	u.data = dup.data
}

func (u Uint) IsZero() bool {
	return u.data[0] == 0 && u.data[1] == 0 && u.data[2] == 0 && u.data[3] == 0
}

func (x Uint) Equals(y *Uint) bool {
	return x.data[0] == y.data[0] && x.data[1] == y.data[1] && x.data[2] == y.data[2] && x.data[3] == y.data[3]
}