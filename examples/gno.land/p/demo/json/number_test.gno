package json

import (
	"testing"
)

func TestParseFloat(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, _ := ParseFloat([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("ParseFloat(%s): got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParseFloatWithScientificNotation(t *testing.T) {
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
			got, err := ParseFloat([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("ParseFloat(%s): got %v, want %v", tc.input, got, tc.expected)
			}

			if err != nil {
				t.Errorf("ParseFloat(%s): got error %v", tc.input, err)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, _ := ParseInt([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("ParseInt(%s): got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}