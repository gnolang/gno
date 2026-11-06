package cid

import "errors"

const (
	base32Lower    = "abcdefghijklmnopqrstuvwxyz234567"
	base32Upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz" // bitcoin
	base36Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	base16Lower    = "0123456789abcdef"
	base16Upper    = "0123456789ABCDEF"
)

var errMultibase = errors.New("cid: invalid multibase string")

// encodeMultibase encodes b in the multibase named by its prefix character.
// The codecs are written here rather than taken from encoding/hex or
// encoding/base32: those are not strict, and loading them would cost every
// transaction that uses this package more gas than the code they replace.
func encodeMultibase(base byte, b []byte) (string, error) {
	switch base {
	case 'b':
		return "b" + encodeBase32(b, base32Lower), nil
	case 'B':
		return "B" + encodeBase32(b, base32Upper), nil
	case 'z':
		return "z" + encodeBaseN(b, base58Alphabet), nil
	case 'k':
		return "k" + encodeBaseN(b, base36Alphabet), nil
	case 'f':
		return "f" + encodeBase16(b, base16Lower), nil
	case 'F':
		return "F" + encodeBase16(b, base16Upper), nil
	}
	return "", errMultibase
}

// decodeMultibase decodes a multibase string. It only accepts the one
// canonical encoding of the bytes: no padding, no whitespace, no case mixing.
func decodeMultibase(s string) ([]byte, error) {
	if s == "" || len(s)-1 > maxEncodedLen(s[0]) {
		return nil, errMultibase
	}

	body := s[1:]
	switch s[0] {
	case 'b':
		return decodeBase32(body, 'a')
	case 'B':
		return decodeBase32(body, 'A')
	case 'z':
		return decodeBaseN(body, 58, base58Digit)
	case 'k':
		return decodeBaseN(body, 36, base36Digit)
	case 'f':
		return decodeBase16(body, 'a')
	case 'F':
		return decodeBase16(body, 'A')
	}
	return nil, errMultibase
}

// maxEncodedLen is the longest body a CID can have in a multibase: the
// length of a maxCIDLen-byte CID in that base. Checking it before decoding
// bounds the quadratic base58btc and base36 decoders.
func maxEncodedLen(base byte) int {
	switch base {
	case 'b', 'B':
		return (maxCIDLen*8 + 4) / 5 // 239
	case 'z':
		return 204 // ceil(149 * log(256) / log(58))
	case 'k':
		return 231 // ceil(149 * log(256) / log(36))
	case 'f', 'F':
		return 2 * maxCIDLen
	}
	return 0
}

// encodeBase32 encodes b as RFC 4648 base32 without padding.
func encodeBase32(b []byte, alphabet string) string {
	out := make([]byte, (len(b)*8+4)/5)
	var acc uint32
	var bits uint
	j := 0
	for _, v := range b {
		acc = acc<<8 | uint32(v)
		bits += 8
		for bits >= 5 {
			bits -= 5
			out[j] = alphabet[(acc>>bits)&31]
			j++
		}
		acc &= 1<<bits - 1
	}
	if bits > 0 {
		out[j] = alphabet[(acc<<(5-bits))&31]
	}
	return string(out)
}

// decodeBase32 decodes unpadded RFC 4648 base32 whose letters start at
// first ('a' or 'A'). It rejects any other character, lengths no encoding
// produces, and non-zero leftover bits, so each byte string has exactly one
// accepted encoding.
func decodeBase32(s string, first byte) ([]byte, error) {
	out := make([]byte, len(s)*5/8)
	var acc uint32
	var bits uint
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d byte
		switch {
		case c-first < 26: // wraps around below first
			d = c - first
		case c-'2' < 6:
			d = c - '2' + 26
		default:
			return nil, errMultibase
		}
		acc = acc<<5 | uint32(d)
		bits += 5
		if bits >= 8 {
			bits -= 8
			out[j] = byte(acc >> bits)
			j++
			acc &= 1<<bits - 1
		}
	}
	if bits >= 5 || acc != 0 {
		return nil, errMultibase
	}
	return out, nil
}

// encodeBase16 encodes b as hex in alphabet's case.
func encodeBase16(b []byte, alphabet string) string {
	out := make([]byte, 2*len(b))
	for i, v := range b {
		out[2*i] = alphabet[v>>4]
		out[2*i+1] = alphabet[v&15]
	}
	return string(out)
}

// decodeBase16 decodes hex whose letters start at first ('a' or 'A'),
// rejecting the other case.
func decodeBase16(s string, first byte) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errMultibase
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d byte
		switch {
		case c-'0' < 10:
			d = c - '0'
		case c-first < 6:
			d = c - first + 10
		default:
			return nil, errMultibase
		}
		out[i/2] = out[i/2]<<4 | d
	}
	return out, nil
}

// encodeBaseN encodes b in the base of alphabet (base58btc, base36): each
// leading zero byte becomes alphabet[0], the rest is converted as one
// number. The conversion is quadratic, so it works on limbs of five digits,
// fed four input bytes at a time: every product stays below 2^63.
func encodeBaseN(b []byte, alphabet string) string {
	base := len(alphabet)
	limbBase := base * base * base * base * base // 58^5 < 2^30
	zeros := 0
	for zeros < len(b) && b[zeros] == 0 {
		zeros++
	}

	var limbs []int // little-endian
	for i := zeros; i < len(b); {
		n := len(b) - i
		if n > 4 {
			n = 4
		}
		carry := 0
		for j := 0; j < n; j++ {
			carry = carry<<8 | int(b[i+j])
		}
		i += n
		mul := 1 << (8 * n)
		for j := 0; j < len(limbs); j++ {
			carry += limbs[j] * mul
			limbs[j] = carry % limbBase
			carry /= limbBase
		}
		for carry > 0 {
			limbs = append(limbs, carry%limbBase)
			carry /= limbBase
		}
	}

	size := zeros
	if len(limbs) > 0 {
		size += 5 * (len(limbs) - 1)
		for top := limbs[len(limbs)-1]; top > 0; top /= base {
			size++
		}
	}
	out := make([]byte, size)
	for i := 0; i < zeros; i++ {
		out[i] = alphabet[0]
	}
	p := size - 1
	for _, v := range limbs {
		for d := 0; d < 5 && p >= zeros; d++ {
			out[p] = alphabet[v%base]
			v /= base
			p--
		}
	}
	return string(out)
}

// decodeBaseN reverses encodeBaseN, reading digits with digit. It works on
// limbs of 32 bits, fed five digits at a time. It is canonical by
// construction: after the leading zero digits, the number has no leading
// zero digit, so no two strings decode to the same bytes.
func decodeBaseN(s string, base int, digit func(byte) int) ([]byte, error) {
	zeros := 0
	for zeros < len(s) && digit(s[zeros]) == 0 {
		zeros++
	}

	var limbs []int // little-endian
	for i := zeros; i < len(s); {
		n := len(s) - i
		if n > 5 {
			n = 5
		}
		carry, mul := 0, 1
		for j := 0; j < n; j++ {
			d := digit(s[i+j])
			if d < 0 {
				return nil, errMultibase
			}
			carry = carry*base + d
			mul *= base
		}
		i += n
		for j := 0; j < len(limbs); j++ {
			carry += limbs[j] * mul
			limbs[j] = carry & 0xffffffff
			carry >>= 32
		}
		for carry > 0 {
			limbs = append(limbs, carry&0xffffffff)
			carry >>= 32
		}
	}

	size := zeros
	if len(limbs) > 0 {
		size += 4 * (len(limbs) - 1)
		for top := limbs[len(limbs)-1]; top > 0; top >>= 8 {
			size++
		}
	}
	out := make([]byte, size)
	p := size - 1
	for _, v := range limbs {
		for d := 0; d < 4 && p >= zeros; d++ {
			out[p] = byte(v)
			v >>= 8
			p--
		}
	}
	return out, nil
}

// Digit decoders for the base58btc and base36 alphabets: each returns the
// value of c, or -1.

func base36Digit(c byte) int {
	switch {
	case c-'0' < 10:
		return int(c - '0')
	case c-'a' < 26:
		return int(c-'a') + 10
	}
	return -1
}

// base58Digit follows the bitcoin alphabet, which skips 0, I, O and l.
func base58Digit(c byte) int {
	switch {
	case '1' <= c && c <= '9':
		return int(c - '1')
	case 'A' <= c && c <= 'H':
		return int(c-'A') + 9
	case 'J' <= c && c <= 'N':
		return int(c-'J') + 17
	case 'P' <= c && c <= 'Z':
		return int(c-'P') + 22
	case 'a' <= c && c <= 'k':
		return int(c-'a') + 33
	case 'm' <= c && c <= 'z':
		return int(c-'m') + 44
	}
	return -1
}
