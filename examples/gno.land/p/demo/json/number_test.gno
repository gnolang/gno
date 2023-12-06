package json

import (
	"testing"
)

func TestParseNumberKind(t *testing.T) {
	cases := []struct {
		name     string
		input    []byte
		expected float64
		isFloat  bool
		ok       bool
		overflow bool
	}{
		{"Positive Integer", []byte("12345"), 12345, false, true, false},
		{"Negative Integer", []byte("-12345"), -12345, false, true, false},
		{"Positive Float", []byte("123.45"), 123.45, true, true, false},
		{"Negative Float", []byte("-123.45"), -123.45, true, true, false},
		{"Integer Overflow", []byte("18446744073709551616"), 0, false, false, true},
		{"Invalid Format", []byte("123a45"), 0, false, false, false},
		{"Multiple Decimal Points", []byte("123.45.67"), 0, false, false, false},
		{"Empty Input", []byte(""), 0, false, false, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			value, isFloat, ok, overflow := ParseNumberKind(c.input)
			if value != c.expected || isFloat != c.isFloat || ok != c.ok || overflow != c.overflow {
				t.Errorf("ParseNumberKind(%s) == (%f, %t, %t, %t), expected (%f, %t, %t, %t)",
					c.input, value, isFloat, ok, overflow, c.expected, c.isFloat, c.ok, c.overflow)
			}
		})
	}
}
