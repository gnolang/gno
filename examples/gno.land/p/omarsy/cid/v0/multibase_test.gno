package cid

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// The multibase spec's test vectors (github.com/multiformats/multibase,
// tests/basic.csv, leading_zero.csv and two_leading_zeros.csv) for the bases
// this package supports.
func TestMultibaseSpecVectors(t *testing.T) {
	for _, tc := range []struct {
		input   string
		encoded []string
	}{
		{"yes mani !", []string{
			"f796573206d616e692021",
			"F796573206D616E692021",
			"bpfsxgidnmfxgsibb",
			"BPFSXGIDNMFXGSIBB",
			"k2lcpzo5yikidynfl",
			"z7paNL19xttacUY",
		}},
		{"\x00yes mani !", []string{
			"f00796573206d616e692021",
			"F00796573206D616E692021",
			"bab4wk4zanvqw42jaee",
			"BAB4WK4ZANVQW42JAEE",
			"k02lcpzo5yikidynfl",
			"z17paNL19xttacUY",
		}},
		{"\x00\x00yes mani !", []string{
			"f0000796573206d616e692021",
			"F0000796573206D616E692021",
			"baaahszltebwwc3tjeaqq",
			"BAAAHSZLTEBWWC3TJEAQQ",
			"k002lcpzo5yikidynfl",
			"z117paNL19xttacUY",
		}},
	} {
		for _, encoded := range tc.encoded {
			got, err := encodeMultibase(encoded[0], []byte(tc.input))
			urequire.NoError(t, err, encoded)
			uassert.Equal(t, encoded, got)

			decoded, err := decodeMultibase(encoded)
			urequire.NoError(t, err, encoded)
			uassert.Equal(t, tc.input, string(decoded))
		}
	}
}

func TestMultibaseRejects(t *testing.T) {
	for _, s := range []string{
		"",
		"x00",
		"bab4wk4zanvqw42jaef", // non-zero leftover bits
		"bpfsxgidnmfxgsibba",  // impossible base32 length
		"bpfsxgidnmfxgsibb=",  // padding
		"bPFSXGIDNMFXGSIBB",   // case mixing
		"k2lcpzo5yikidynfL",   // case mixing
		"z7paNL19xttacU0",     // not in base58btc
		"f796573206d616e69202",
		"f796573206D616E692021",
		"F796573206d616e692021",
		// Longer than any CID: rejected before decoding.
		"b" + strings.Repeat("a", 240),
		"z" + strings.Repeat("2", 205),
		"k" + strings.Repeat("2", 232),
		"f" + strings.Repeat("0", 300),
	} {
		_, err := decodeMultibase(s)
		uassert.Error(t, err, s)
	}
}

// Every byte value decodes to its index in the alphabet, or is rejected.
func TestDigits(t *testing.T) {
	for c := 0; c < 256; c++ {
		b := byte(c)

		// Base32 and base16 decode digits inline: probe them with the digit
		// followed by the zero digit.
		for _, tc := range []struct {
			alphabet string
			decode   func(string) ([]byte, error)
			shift    uint
		}{
			{base32Lower, func(s string) ([]byte, error) { return decodeBase32(s, 'a') }, 3},
			{base32Upper, func(s string) ([]byte, error) { return decodeBase32(s, 'A') }, 3},
			{base16Lower, func(s string) ([]byte, error) { return decodeBase16(s, 'a') }, 4},
			{base16Upper, func(s string) ([]byte, error) { return decodeBase16(s, 'A') }, 4},
		} {
			got, err := tc.decode(string([]byte{b, tc.alphabet[0]}))
			if want := strings.IndexByte(tc.alphabet, b); want < 0 {
				uassert.Error(t, err, tc.alphabet)
			} else {
				urequire.NoError(t, err, tc.alphabet)
				uassert.Equal(t, int(byte(want<<tc.shift)), int(got[0]), tc.alphabet)
			}
		}

		uassert.Equal(t, strings.IndexByte(base36Alphabet, b), base36Digit(b))
		uassert.Equal(t, strings.IndexByte(base58Alphabet, b), base58Digit(b))
	}
}

// encodeBaseN and decodeBaseN work on multi-digit limbs; check them against
// the plain digit-by-digit conversion on inputs of every length up to the
// longest CID, with and without leading zeros.
func TestBaseNLimbs(t *testing.T) {
	seed := uint32(1)
	for size := 0; size <= maxCIDLen; size++ {
		b := make([]byte, size)
		for i := range b {
			seed = seed*1664525 + 1013904223
			b[i] = byte(seed >> 24)
		}
		if size%3 == 0 && size > 2 {
			b[0], b[1] = 0, 0
		}

		for _, alphabet := range []string{base58Alphabet, base36Alphabet} {
			enc := encodeBaseN(b, alphabet)
			uassert.Equal(t, refEncodeBaseN(b, alphabet), enc)

			digit := base58Digit
			if alphabet == base36Alphabet {
				digit = base36Digit
			}
			dec, err := decodeBaseN(enc, len(alphabet), digit)
			urequire.NoError(t, err)
			uassert.Equal(t, string(b), string(dec))
		}
	}
}

// refEncodeBaseN is the plain repeated-division conversion.
func refEncodeBaseN(b []byte, alphabet string) string {
	base := len(alphabet)
	zeros := 0
	for zeros < len(b) && b[zeros] == 0 {
		zeros++
	}
	var digits []byte // little-endian
	for i := zeros; i < len(b); i++ {
		carry := int(b[i])
		for j := 0; j < len(digits); j++ {
			carry += int(digits[j]) << 8
			digits[j] = byte(carry % base)
			carry /= base
		}
		for carry > 0 {
			digits = append(digits, byte(carry%base))
			carry /= base
		}
	}
	out := strings.Repeat(alphabet[:1], zeros)
	for i := len(digits) - 1; i >= 0; i-- {
		out += alphabet[digits[i] : digits[i]+1]
	}
	return out
}
