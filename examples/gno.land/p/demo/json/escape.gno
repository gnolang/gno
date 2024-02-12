package json

import (
	"bytes"
	"errors"
	"unicode/utf8"
)

const (
	supplementalPlanesOffset = 0x10000
	highSurrogateOffset      = 0xD800
	lowSurrogateOffset       = 0xDC00

	surrogateEnd                 = 0xDFFF
	basicMultilingualPlaneOffset = 0xFFFF

	badHex = -1
)

// unescape takes an input byte slice, processes it to unescape certain characters,
// and writes the result into an output byte slice.
//
// it returns the processed slice and any error encountered during the unescape operation.
func unescape(input, output []byte) ([]byte, error) {
	// find the index of the first backslash in the input slice.
	firstBackslash := bytes.IndexByte(input, BackSlashToken)
	if firstBackslash == -1 {
		return input, nil
	}

	// ensure the output slice has enough capacity to hold the result.
	inputLen := len(input)
	if cap(output) < inputLen {
		output = make([]byte, inputLen)
	} else {
		// if the capacity is sufficient, slice the output to the length of the input.
		output = output[:inputLen]
	}

	copy(output, input[:firstBackslash])
	input = input[firstBackslash:]
	buf := output[firstBackslash:]

	for len(input) > 0 {
		inLen, bufLen := processEscapedUTF8(input, buf)
		if inLen == -1 {
			return nil, errors.New("JSON Error: Encountered an invalid escape sequence in a string.")
		}

		input = input[inLen:]	// the number of bytes consumed in the input
		buf = buf[bufLen:]		// the number of bytes written to buf

		// find the next backslash in the remaining input
		nextBackslash := bytes.IndexByte(input, BackSlashToken)
		if nextBackslash == -1 {
			copy(buf, input)
			buf = buf[len(input):]

			break
		}

		copy(buf, input[:nextBackslash])

		input = input[nextBackslash:]
		buf = buf[nextBackslash:]
	}

	return output[:len(output)-len(buf)], nil
}

// isSurrogatePair returns true if the rune is a surrogate pair.
//
// A surrogate pairs are used in UTF-16 encoding to encode characters
// outside the Basic Multilingual Plane (BMP).
func isSurrogatePair(r rune) bool {
	return highSurrogateOffset <= r && r <= surrogateEnd
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
func decodeUnicodeEscape(b []byte) (rune, int) {
	r, ok := decodeSingleUnicodeEscape(b)
	if !ok {
		return utf8.RuneError, -1
	}

	// determine valid unicode escapes within the BMP
	if r <= basicMultilingualPlaneOffset && !isSurrogatePair(r) {
		return r, 6
	}

	// Decode the following escape sequence to verify a UTF-16 susergate pair.
	r2, ok := decodeSingleUnicodeEscape(b[6:])
	if !ok {
		return utf8.RuneError, -1
	}

	if r2 < lowSurrogateOffset {
		return utf8.RuneError, -1
	}

	return combineSurrogates(r, r2), 12
}

var escapeByteSet = [256]byte{
	'"':  DoublyQuoteToken,
	'\\': BackSlashToken,
	'/':  SlashToken,
	'b':  BackSpaceToken,
	'f':  FormFeedToken,
	'n':  NewLineToken,
	'r':  CarriageReturnToken,
	't':  TabToken,
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
	if len(in) < 2 || in[0] != BackSlashToken {
		return -1, -1
	}

	escapeSeqLen := 2
	escapeChar := in[1]
	if escapeChar == 'u' {
		if r, size := decodeUnicodeEscape(in); size != -1 {
			outLen = utf8.EncodeRune(out, r)
			return size, outLen
		}
	} else {
		val := escapeByteSet[escapeChar]
		if val != 0 {
			out[0] = val
			return escapeSeqLen, 1
		}
	}

	return -1, -1
}
