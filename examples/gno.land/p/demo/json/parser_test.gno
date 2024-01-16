package json

import (
	"bytes"
	"testing"
)

func TestParseStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isError  bool
	}{
		{`"Hello, World!"`, "\"Hello, World!\"", false},
		{`\uFF11`, "\uFF11", false},
		{`\uFFFF`, "\uFFFF", false},
		{`true`, "true", false},
		{`false`, "false", false},
		{`\uDF00`, "", true},
	}

	for i, tt := range tests {
		s, err := ParseStringLiteral([]byte(tt.input))

		if !tt.isError && err != nil {
			t.Errorf("%d. unexpected error: %s", i, err)
		}

		if tt.isError && err == nil {
			t.Errorf("%d. expected error, but not error", i)
		}

		if s != tt.expected {
			t.Errorf("%d. expected=%s, but actual=%s", i, tt.expected, s)
		}
	}
}

func TestParseBoolLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		isError  bool
	}{
		{`true`, true, false},
		{`false`, false, false},
		{`TRUE`, false, true},
		{`FALSE`, false, true},
		{`foo`, false, true},
		{`"true"`, false, true},
		{`"false"`, false, true},
	}

	for i, tt := range tests {
		b, err := ParseBoolLiteral([]byte(tt.input))

		if !tt.isError && err != nil {
			t.Errorf("%d. unexpected error: %s", i, err)
		}

		if tt.isError && err == nil {
			t.Errorf("%d. expected error, but not error", i)
		}

		if b != tt.expected {
			t.Errorf("%d. expected=%t, but actual=%t", i, tt.expected, b)
		}
	}
}

func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123", 123},
		{"-123", -123},
		{"123.456", 123.456},
		{"-123.456", -123.456},
		{"0.123", 0.123},
		{"-0.123", -0.123},
		{"", -1},
		{"abc", -1},
		{"123.45.6", -1},
		{"999999999999999999999", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFloatWithScientificNotation(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1e6", 1000000},
		{"1E6", 1000000},
		{"1.23e10", 1.23e10},
		{"1.23E10", 1.23e10},
		{"-1.23e10", -1.23e10},
		{"-1.23E10", -1.23e10},
		{"2.45e-8", 2.45e-8},
		{"2.45E-8", 2.45e-8},
		{"-2.45e-8", -2.45e-8},
		{"-2.45E-8", -2.45e-8},
		{"5e0", 5},
		{"-5e0", -5},
		{"5E+0", 5},
		{"5e+1", 50},
		{"1e-1", 0.1},
		{"1E-1", 0.1},
		{"-1e-1", -0.1},
		{"-1E-1", -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFloatLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseFloatLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}

			if err != nil {
				t.Errorf("ParseFloatLiteral(%s): got error %v", tt.input, err)
			}
		})
	}
}

func TestParseIntLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
		{"12345", 12345},
		{"-12345", -12345},
		{"9223372036854775807", 9223372036854775807},
		{"-9223372036854775808", -9223372036854775808},
		{"-92233720368547758081", 0},
		{"18446744073709551616", 0},
		{"9223372036854775808", 0},
		{"-9223372036854775809", 0},
		{"", 0},
		{"abc", 0},
		{"12345x", 0},
		{"123e5", 0},
		{"9223372036854775807x", 0},
		{"27670116110564327410", 0},
		{"-27670116110564327410", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := ParseIntLiteral([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ParseIntLiteral(%s): got %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

type parseUint64Test struct {
	in  string
	out uint64
	err bool
}

var parseUint64Tests = []parseUint64Test{
	{"", 0, true},
	{"0", 0, nil},
	{"1", 1, nil},
	{"12345", 12345, nil},
	{"012345", 12345, nil},
	{"12345x", 0, true},
	{"98765432100", 98765432100, nil},
	{"18446744073709551615", 1<<64 - 1, nil},
	{"18446744073709551616", 1<<64 - 1, true},
	{"18446744073709551620", 1<<64 - 1, true},
	{"1_2_3_4_5", 0, true}, // base=10 so no underscores allowed
	{"_12345", 0, true},
	{"1__2345", 0, true},
	{"12345_", 0, true},
	{"-0", 0, true},
	{"-1", 0, true},
	{"+1", 0, true},
}

func TestGetType(t *testing.T) {
	tests := []struct {
		data     []byte
		offset   int
		expected []byte
		vt       ValueType
		err      error
	}{
		{[]byte(`{"name": "John", "age": 30}`), 0, []byte(`{"name": "John", "age": 30}`), Object, nil},
		{[]byte(`"Hello, World!"`), 0, []byte(`"Hello, World!"`), String, nil},
		{[]byte(`12345`), 0, []byte(`12345`), Number, nil},
		{[]byte(`true`), 0, []byte(`true`), Boolean, nil},
		{[]byte(`null`), 0, []byte(`null`), Null, nil},
		{[]byte(`[1, 2, 3]`), 0, []byte(`[1, 2, 3]`), Array, nil},
		{[]byte(`{}`), 0, []byte(`{}`), Object, nil},
	}

	for i, tt := range tests {
		gotData, gotVT, _, gotErr := getTypeFromByteSlice(tt.data, tt.offset)

		if !bytes.Equal(gotData, tt.expected) {
			t.Errorf("%d. expected data=%s, but got data=%s", i, tt.expected, gotData)
		}

		if gotVT != tt.vt {
			t.Errorf("%d. expected value type=%s, but got value type=%s", i, tt.vt, gotVT)
		}

		if gotErr != tt.err {
			t.Errorf("%d. expected error=%v, but got error=%v", i, tt.err, gotErr)
		}
	}
}
