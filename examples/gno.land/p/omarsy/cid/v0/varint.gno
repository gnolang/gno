package cid

import "errors"

const (
	// maxVarintLen is the longest unsigned varint the multiformats spec
	// allows: 9 bytes, carrying 63 bits.
	maxVarintLen = 9
	maxVarint    = 1<<63 - 1
)

var errVarint = errors.New("cid: invalid varint")

// uvarint decodes the unsigned varint at the start of s and returns it with
// its length. It rejects encodings longer than 9 bytes and non-minimal ones
// (a final zero byte after the first), so every value has one encoding.
// encoding/binary.Uvarint accepts both, and wrapping it would also cost
// every transaction the gas of loading that package.
func uvarint(s string) (uint64, int, error) {
	var x uint64
	for i := 0; i < maxVarintLen && i < len(s); i++ {
		c := s[i]
		x |= uint64(c&0x7f) << (7 * uint(i))
		if c < 0x80 {
			if c == 0 && i > 0 {
				return 0, 0, errVarint
			}
			return x, i + 1, nil
		}
	}
	return 0, 0, errVarint
}

// appendUvarint appends the minimal unsigned varint encoding of x, which
// must be at most maxVarint, to b.
func appendUvarint(b []byte, x uint64) []byte {
	for x >= 0x80 {
		b = append(b, byte(x)|0x80)
		x >>= 7
	}
	return append(b, byte(x))
}
