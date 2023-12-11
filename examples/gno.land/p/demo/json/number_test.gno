package json

import (
	"testing"
)

func TestParseNumberKind(t *testing.T) {
	testCases := []struct {
		input         string
		expected      float64
		expectedPanic bool
	}{
		{"123", 123, false},
		{"-123", -123, false},
		{"123.456", 123.456, false},
		{"-123.456", -123.456, false},
		{"0.123", 0.123, false},
		{"-0.123", -0.123, false},
		{"", 0, true},
		{"abc", 0, true},
		{"123.45.6", 0, true},
		{"999999999999999999999", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tc.expectedPanic {
					t.Errorf("ParseNumberKind(%s): expected panic=%v, got panic=%v",
						tc.input, tc.expectedPanic, r != nil)
				}
			}()

			got, _, _ := ParseNumberKind([]byte(tc.input))
			if got != tc.expected && !tc.expectedPanic {
				t.Errorf("ParseNumberKind(%s): expected %v, got %v",
					tc.input, tc.expected, got)
			}
		})
	}
}

func TestParseNumberKindWithScientificNotation(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, _, _ := ParseNumberKind([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("ParseNumberKind(%s): got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
