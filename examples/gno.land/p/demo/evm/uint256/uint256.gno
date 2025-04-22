package uint256

import (
	"strconv"
	"strings"
	"errors"
)

type uint256 struct {
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

func NewUint256FromString(s string, base int) (*uint256, error) {
	u := NewUint256()
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

func NewUint256() *uint256 {
	return &uint256{data: [4]uint64{0, 0, 0, 0}}
}

// string

func (u uint256) HexString() string {
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

func (u uint256) divMod10() (uint256, uint64) {
	var result uint256
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

func (u uint256) String() string {
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

func (u *uint256) add64Level(x uint64, n int) {
	for i := n; true; i++ {
		res, carry := carryAdd(u.data[i], x)
		u.data[i] = res
		if carry == 0 {
			break
		}
		x = carry
	}
}

func (u *uint256) Add64(x uint64) {
	u.add64Level(x, 0)
}

func (u *uint256) Mul64(x uint64) {
	dup := NewUint256()
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

func (u uint256) IsZero() bool {
	return u.data[0] == 0 && u.data[1] == 0 && u.data[2] == 0 && u.data[3] == 0
}