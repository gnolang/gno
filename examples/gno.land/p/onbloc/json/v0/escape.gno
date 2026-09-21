package json

import (
	"unicode/utf8"
)

const (
	supplementalPlanesOffset     = 0x10000
	highSurrogateOffset          = 0xD800
	lowSurrogateOffset           = 0xDC00
	surrogateEnd                 = 0xDFFF
	basicMultilingualPlaneOffset = 0xFFFF
	badHex                       = -1

	singleUnicodeEscapeLen = 6
	surrogatePairLen       = 12
)

var hexLookupTable = [256]int{
	'0': 0x0, '1': 0x1, '2': 0x2, '3': 0x3, '4': 0x4,
	'5': 0x5, '6': 0x6, '7': 0x7, '8': 0x8, '9': 0x9,
	'A': 0xA, 'B': 0xB, 'C': 0xC, 'D': 0xD, 'E': 0xE, 'F': 0xF,
	'a': 0xA, 'b': 0xB, 'c': 0xC, 'd': 0xD, 'e': 0xE, 'f': 0xF,
	// Fill unspecified index-value pairs with key and value of -1
	'G': -1, 'H': -1, 'I': -1, 'J': -1,
	'K': -1, 'L': -1, 'M': -1, 'N': -1,
	'O': -1, 'P': -1, 'Q': -1, 'R': -1,
	'S': -1, 'T': -1, 'U': -1, 'V': -1,
	'W': -1, 'X': -1, 'Y': -1, 'Z': -1,
	'g': -1, 'h': -1, 'i': -1, 'j': -1,
	'k': -1, 'l': -1, 'm': -1, 'n': -1,
	'o': -1, 'p': -1, 'q': -1, 'r': -1,
	's': -1, 't': -1, 'u': -1, 'v': -1,
	'w': -1, 'x': -1, 'y': -1, 'z': -1,
}

func h2i(c byte) int {
	return hexLookupTable[c]
}

// Unescape takes an input byte slice, processes it to Unescape certain characters,
// and writes the result into an output byte slice.
//
// it returns the processed slice and any error encountered during the Unescape operation.
func Unescape(input, output []byte) ([]byte, error) {
	// ensure the output slice has enough capacity to hold the input slice.
	inputLen := len(input)
	if cap(output) < inputLen {
		output = make([]byte, inputLen)
	}

	inPos, outPos := 0, 0

	for inPos < len(input) {
		c := input[inPos]
		if c != backSlash {
			output[outPos] = c
			inPos++
			outPos++
		} else {
			// process escape sequence
			inLen, outLen, err := processEscapedUTF8(input[inPos:], output[outPos:])
			if err != nil {
				return nil, err
			}
			inPos += inLen
			outPos += outLen
		}
	}

	return output[:outPos], nil
}

// isSurrogatePair returns true if the rune is a surrogate pair.
//
// A surrogate pairs are used in UTF-16 encoding to encode characters
// outside the Basic Multilingual Plane (BMP).
func isSurrogatePair(r rune) bool {
	return highSurrogateOffset <= r && r <= surrogateEnd
}

// isHighSurrogate checks if the rune is a high surrogate (U+D800 to U+DBFF).
func isHighSurrogate(r rune) bool {
	return r >= highSurrogateOffset && r <= 0xDBFF
}

// isLowSurrogate checks if the rune is a low surrogate (U+DC00 to U+DFFF).
func isLowSurrogate(r rune) bool {
	return r >= lowSurrogateOffset && r <= surrogateEnd
}

// combineSurrogates reconstruct the original unicode code points in the
// supplemental plane by combinin the high and low surrogate.
//
// The hight surrogate in the range from U+D800 to U+DBFF,
// and the low surrogate in the range from U+DC00 to U+DFFF.
//
// The formula to combine the surrogates is:
// (high - 0xD800) * 0x400 + (low - 0xDC00) + 0x10000
func combineSurrogates(high, low rune) rune {
	return ((high - highSurrogateOffset) << 10) + (low - lowSurrogateOffset) + supplementalPlanesOffset
}

// deocdeSingleUnicodeEscape decodes a unicode escape sequence (e.g., \uXXXX) into a rune.
func decodeSingleUnicodeEscape(b []byte) (rune, bool) {
	if len(b) < 6 {
		return utf8.RuneError, false
	}

	// convert hex to decimal
	h1, h2, h3, h4 := h2i(b[2]), h2i(b[3]), h2i(b[4]), h2i(b[5])
	if h1 == badHex || h2 == badHex || h3 == badHex || h4 == badHex {
		return utf8.RuneError, false
	}

	return rune(h1<<12 + h2<<8 + h3<<4 + h4), true
}

// decodeUnicodeEscape decodes a Unicode escape sequence from a byte slice.
// It handles both single Unicode escape sequences and surrogate pairs.
func decodeUnicodeEscape(b []byte) (rune, int) {
	// decode the first Unicode escape sequence.
	r, ok := decodeSingleUnicodeEscape(b)
	if !ok {
		return utf8.RuneError, -1
	}

	// if the rune is within the BMP and not a surrogate, return it
	if r <= basicMultilingualPlaneOffset && !isSurrogatePair(r) {
		return r, 6
	}

	if !isHighSurrogate(r) {
		// invalid surrogate pair.
		return utf8.RuneError, -1
	}

	// if the rune is a high surrogate, need to decode the next escape sequence.

	// ensure there are enough bytes for the next escape sequence.
	if len(b) < surrogatePairLen {
		return utf8.RuneError, -1
	}
	// decode the second Unicode escape sequence.
	r2, ok := decodeSingleUnicodeEscape(b[singleUnicodeEscapeLen:])
	if !ok {
		return utf8.RuneError, -1
	}
	// check if the second rune is a low surrogate.
	if isLowSurrogate(r2) {
		combined := combineSurrogates(r, r2)
		return combined, surrogatePairLen
	}
	return utf8.RuneError, -1
}

var escapeByteSet = [256]byte{
	'"':  doubleQuote,
	'\\': backSlash,
	'/':  slash,
	'b':  backSpace,
	'f':  formFeed,
	'n':  newLine,
	'r':  carriageReturn,
	't':  tab,
}

// Unquote takes a byte slice and unquotes it by removing
// the surrounding quotes and unescaping the contents.
func Unquote(s []byte, border byte) (string, bool) {
	s, ok := unquoteBytes(s, border)
	return string(s), ok
}

// unquoteBytes takes a byte slice and unquotes it by removing
func unquoteBytes(s []byte, border byte) ([]byte, bool) {
	if len(s) < 2 || s[0] != border || s[len(s)-1] != border {
		return nil, false
	}

	s = s[1 : len(s)-1]

	r := 0
	for r < len(s) {
		c := s[r]

		if c == backSlash || c == border || c < 0x20 {
			break
		}

		if c < utf8.RuneSelf {
			r++
			continue
		}

		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}

		r += size
	}

	if r == len(s) {
		return s, true
	}

	utfDoubleMax := utf8.UTFMax * 2
	b := make([]byte, len(s)+utfDoubleMax)
	w := copy(b, s[0:r])

	for r < len(s) {
		if w >= len(b)-utf8.UTFMax {
			nb := make([]byte, utfDoubleMax+(2*len(b)))
			copy(nb, b)
			b = nb
		}

		c := s[r]
		if c == backSlash {
			r++
			if r >= len(s) {
				return nil, false
			}

			if s[r] == 'u' {
				rr, res := decodeUnicodeEscape(s[r-1:])
				if res < 0 {
					return nil, false
				}

				w += utf8.EncodeRune(b[w:], rr)
				r += 5
			} else {
				decode := escapeByteSet[s[r]]
				if decode == 0 {
					return nil, false
				}

				if decode == doubleQuote || decode == backSlash || decode == slash {
					decode = s[r]
				}

				b[w] = decode
				r++
				w++
			}
		} else if c == border || c < 0x20 {
			return nil, false
		} else if c < utf8.RuneSelf {
			b[w] = c
			r++
			w++
		} else {
			rr, size := utf8.DecodeRune(s[r:])

			if rr == utf8.RuneError && size == 1 {
				return nil, false
			}

			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}

	return b[:w], true
}

// processEscapedUTF8 converts escape sequences to UTF-8 characters.
// It decodes Unicode escape sequences (\uXXXX) to UTF-8 and
// converts standard escape sequences (e.g., \n) to their corresponding special characters.
func processEscapedUTF8(in, out []byte) (int, int, error) {
	if len(in) < 2 || in[0] != backSlash {
		return -1, -1, errInvalidEscapeSequence
	}

	escapeSeqLen := 2
	escapeChar := in[1]

	if escapeChar != 'u' {
		val := escapeByteSet[escapeChar]
		if val == 0 {
			return -1, -1, errInvalidEscapeSequence
		}

		out[0] = val
		return escapeSeqLen, 1, nil
	}

	r, size := decodeUnicodeEscape(in)
	if size == -1 {
		return -1, -1, errInvalidEscapeSequence
	}

	outLen := utf8.EncodeRune(out, r)

	return size, outLen, nil
}
