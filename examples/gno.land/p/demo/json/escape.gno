package json

import (
	"bytes"
	"unicode/utf8"
)

const (
	SupplementalPlanesOffset = 0x10000
	HighSurrogateOffset      = 0xD800
	LowSurrogateOffset       = 0xDC00

	SurrogateEnd                 = 0xDFFF
	BasicMultilingualPlaneOffset = 0xFFFF

	BadHex = -1
)

var escapeMap = map[byte]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

// hexToInt converts a hex character to its integer value.
func hexToInt(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	}

	return BadHex
}

// isSurrogatePair returns true if the rune is a surrogate pair.
//
// A surrogate pairs are used in UTF-16 encoding to encode characters
// outside the Basic Multilingual Plane (BMP).
func isSurrogatePair(r rune) bool {
	return HighSurrogateOffset <= r && r <= SurrogateEnd
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
	return ((high - HighSurrogateOffset) << 10) + (low - LowSurrogateOffset) + SupplementalPlanesOffset
}

// deocdeSingleUnicodeEscape decodes a unicode escape sequence (e.g., \uXXXX) into a rune.
func decodeSingleUnicodeEscape(b []byte) (rune, bool) {
	if len(b) < 6 {
		return utf8.RuneError, false
	}

	// convert hex to decimal
	h1, h2, h3, h4 := hexToInt(b[2]), hexToInt(b[3]), hexToInt(b[4]), hexToInt(b[5])
	if h1 == BadHex || h2 == BadHex || h3 == BadHex || h4 == BadHex {
		return utf8.RuneError, false
	}

	return rune(h1<<12 + h2<<8 + h3<<4 + h4), true
}

// decodeUnicodeEscape decodes a Unicode escape sequence from a byte slice.
func decodeUnicodeEscape(b []byte) (rune, int) {
	r, ok := decodeSingleUnicodeEscape(b)
	if !ok {
		return utf8.RuneError, -1
	}

	// determine valid unicode escapes within the BMP
	if r <= BasicMultilingualPlaneOffset && !isSurrogatePair(r) {
		return r, 6
	}

	// Decode the following escape sequence to verify a UTF-16 susergate pair.
	r2, ok := decodeSingleUnicodeEscape(b[6:])
	if !ok {
		return utf8.RuneError, -1
	}

	if r2 < LowSurrogateOffset {
		return utf8.RuneError, -1
	}

	return combineSurrogates(r, r2), 12
}

// processEscapedUTF8 processes the escape sequence in the given byte slice and
// and converts them to UTF-8 characters. The function returns the length of the processed input and output.
//
// The input 'in' must contain the escape sequence to be processed,
// and 'out' provides a space to store the converted characters.
//
// The function returns (input length, output length) if the escape sequence is correct.
// Unicode escape sequences (e.g. \uXXXX) are decoded to UTF-8, other default escape sequences are
// converted to their corresponding special characters (e.g. \n -> newline).
//
// If the escape sequence is invalid, or if 'in' does not completely enclose the escape sequence,
// function returns (-1, -1) to indicate an error.
func processEscapedUTF8(in, out []byte) (inLen int, outLen int) {
	if len(in) < 2 || in[0] != '\\' {
		return -1, -1
	}

	if val, ok := escapeMap[in[1]]; ok {
		out[0] = val
		return 2, 1
	}

	if in[1] == 'u' {
		if r, size := decodeUnicodeEscape(in); size != -1 {
			outLen = utf8.EncodeRune(out, r)
			return size, outLen
		}
	}

	return -1, -1
}

func Unescape(input, output []byte) []byte {
	firstBackslash := bytes.IndexByte(input, '\\')
	if firstBackslash == -1 {
		return input
	}

	if cap(output) < len(input) {
		output = make([]byte, len(input))
	} else {
		output = output[:len(input)]
	}

	copy(output, input[:firstBackslash])
	input = input[firstBackslash:]
	buf := output[firstBackslash:]

	for len(input) > 0 {
		inLen, bufLen := processEscapedUTF8(input, buf)
		if inLen == -1 {
			return nil
		}

		input = input[inLen:]
		buf = buf[bufLen:]

		// copy everything until the next backslash
		nextBackslash := bytes.IndexByte(input, '\\')
		if nextBackslash == -1 {
			copy(buf, input)
			buf = buf[len(input):]
			break
		} else {
			copy(buf, input[:nextBackslash])
			buf = buf[nextBackslash:]
			input = input[nextBackslash:]
		}
	}

	return output[:len(output)-len(buf)]
}
