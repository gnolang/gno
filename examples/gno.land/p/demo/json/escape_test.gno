package json

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestHexToInt(t *testing.T) {
	tests := []struct {
		name string
		c    byte
		want int
	}{
		{"Digit 0", '0', 0},
		{"Digit 9", '9', 9},
		{"Uppercase A", 'A', 10},
		{"Uppercase F", 'F', 15},
		{"Lowercase a", 'a', 10},
		{"Lowercase f", 'f', 15},
		{"Invalid character1", 'g', badHex},
		{"Invalid character2", 'G', badHex},
		{"Invalid character3", 'z', badHex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := h2i(tt.c); got != tt.want {
				t.Errorf("h2i() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSurrogatePair(t *testing.T) {
	testCases := []struct {
		name     string
		r        rune
		expected bool
	}{
		{"high surrogate start", 0xD800, true},
		{"high surrogate end", 0xDBFF, true},
		{"low surrogate start", 0xDC00, true},
		{"low surrogate end", 0xDFFF, true},
		{"Non-surrogate", 0x0000, false},
		{"Non-surrogate 2", 0xE000, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSurrogatePair(tc.r); got != tc.expected {
				t.Errorf("isSurrogate() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestCombineSurrogates(t *testing.T) {
	testCases := []struct {
		high, low rune
		expected  rune
	}{
		{0xD83D, 0xDC36, 0x1F436}, // 🐶 U+1F436 DOG FACE
		{0xD83D, 0xDE00, 0x1F600}, // 😀 U+1F600 GRINNING FACE
		{0xD83C, 0xDF03, 0x1F303}, // 🌃 U+1F303 NIGHT WITH STARS
	}

	for _, tc := range testCases {
		result := combineSurrogates(tc.high, tc.low)
		if result != tc.expected {
			t.Errorf("combineSurrogates(%U, %U) = %U; want %U", tc.high, tc.low, result, tc.expected)
		}
	}
}

func TestDecodeSingleUnicodeEscape(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected rune
		isValid  bool
	}{
		// valid unicode escape sequences
		{[]byte(`\u0041`), 'A', true},
		{[]byte(`\u03B1`), 'α', true},
		{[]byte(`\u00E9`), 'é', true}, // valid non-English character
		{[]byte(`\u0021`), '!', true}, // valid special character
		{[]byte(`\uFF11`), '１', true},
		{[]byte(`\uD83D`), 0xD83D, true},
		{[]byte(`\uDE03`), 0xDE03, true},

		// invalid unicode escape sequences
		{[]byte(`\u004`), utf8.RuneError, false},  // too short
		{[]byte(`\uXYZW`), utf8.RuneError, false}, // invalid hex
		{[]byte(`\u00G1`), utf8.RuneError, false}, // non-hex character
	}

	for _, tc := range testCases {
		result, isValid := decodeSingleUnicodeEscape(tc.input)
		if result != tc.expected || isValid != tc.isValid {
			t.Errorf("decodeSingleUnicodeEscape(%s) = (%U, %v); want (%U, %v)", tc.input, result, isValid, tc.expected, tc.isValid)
		}
	}
}

func TestDecodeUnicodeEscape(t *testing.T) {
	testCases := []struct {
		input    string
		expected rune
		size     int
	}{
		{"\\u0041", 'A', 6},
		{"\\u03B1", 'α', 6},
		{"\\u1F600", 0x1F60, 6},
		{"\\uD830\\uDE03", 0x1C203, 12},
		{"\\uD800\\uDC00", 0x00010000, 12},

		{"\\u004", utf8.RuneError, -1},
		{"\\uXYZW", utf8.RuneError, -1},
		{"\\uD83D\\u0041", utf8.RuneError, -1},
	}

	for _, tc := range testCases {
		r, size := decodeUnicodeEscape([]byte(tc.input))
		if r != tc.expected || size != tc.size {
			t.Errorf("decodeUnicodeEscape(%q) = (%U, %d); want (%U, %d)", tc.input, r, size, tc.expected, tc.size)
		}
	}
}

func TestUnescapeToUTF8(t *testing.T) {
	testCases := []struct {
		input       []byte
		expectedIn  int
		expectedOut int
		isError     bool
	}{
		// valid escape sequences
		{[]byte(`\n`), 2, 1, false},
		{[]byte(`\t`), 2, 1, false},
		{[]byte(`\u0041`), 6, 1, false},
		{[]byte(`\u03B1`), 6, 2, false},
		{[]byte(`\uD830\uDE03`), 12, 4, false},

		// invalid escape sequences
		{[]byte(`\`), -1, -1, true},            // incomplete escape sequence
		{[]byte(`\x`), -1, -1, true},           // invalid escape character
		{[]byte(`\u`), -1, -1, true},           // incomplete unicode escape sequence
		{[]byte(`\u004`), -1, -1, true},        // invalid unicode escape sequence
		{[]byte(`\uXYZW`), -1, -1, true},       // invalid unicode escape sequence
		{[]byte(`\uD83D\u0041`), -1, -1, true}, // invalid unicode escape sequence
	}

	for _, tc := range testCases {
		input := make([]byte, len(tc.input))
		copy(input, tc.input)
		output := make([]byte, utf8.UTFMax)
		inLen, outLen, err := processEscapedUTF8(input, output)
		if (err != nil) != tc.isError {
			t.Errorf("processEscapedUTF8(%q) = %v; want %v", tc.input, err, tc.isError)
		}

		if inLen != tc.expectedIn || outLen != tc.expectedOut {
			t.Errorf("processEscapedUTF8(%q) = (%d, %d); want (%d, %d)", tc.input, inLen, outLen, tc.expectedIn, tc.expectedOut)
		}
	}
}

func TestUnescape(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"NoEscape", []byte("hello world"), []byte("hello world")},
		{"SingleEscape", []byte("hello\\nworld"), []byte("hello\nworld")},
		{"MultipleEscapes", []byte("line1\\nline2\\r\\nline3"), []byte("line1\nline2\r\nline3")},
		{"UnicodeEscape", []byte("snowman:\\u2603"), []byte("snowman:\u2603")},
		{"Complex", []byte("tc\\n\\u2603\\r\\nend"), []byte("tc\n\u2603\r\nend")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := Unescape(tc.input, make([]byte, len(tc.input)+10))
			if !bytes.Equal(output, tc.expected) {
				t.Errorf("unescape(%q) = %q; want %q", tc.input, output, tc.expected)
			}
		})
	}
}

func TestUnquoteBytes(t *testing.T) {
	tests := []struct {
		input    []byte
		border   byte
		expected []byte
		ok       bool
	}{
		{[]byte("\"hello\""), '"', []byte("hello"), true},
		{[]byte("'hello'"), '\'', []byte("hello"), true},
		{[]byte("\"hello"), '"', nil, false},
		{[]byte("hello\""), '"', nil, false},
		{[]byte("\"he\\\"llo\""), '"', []byte("he\"llo"), true},
		{[]byte("\"he\\nllo\""), '"', []byte("he\nllo"), true},
		{[]byte("\"\""), '"', []byte(""), true},
		{[]byte("''"), '\'', []byte(""), true},
		{[]byte("\"\\u0041\""), '"', []byte("A"), true},
		{[]byte(`"Hello, 世界"`), '"', []byte("Hello, 世界"), true},
		{[]byte(`"Hello, \x80"`), '"', nil, false},
	}

	for _, tc := range tests {
		result, pass := unquoteBytes(tc.input, tc.border)

		if pass != tc.ok {
			t.Errorf("unquoteBytes(%q) = %v; want %v", tc.input, pass, tc.ok)
		}

		if !bytes.Equal(result, tc.expected) {
			t.Errorf("unquoteBytes(%q) = %q; want %q", tc.input, result, tc.expected)
		}
	}
}
