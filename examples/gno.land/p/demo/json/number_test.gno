package json

import (
	"testing"
)

func TestParseFloat(t *testing.T) {
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
					t.Errorf("ParseFloat(%s): expected panic=%v, got panic=%v",
						tc.input, tc.expectedPanic, r != nil)
				}
			}()

			got, _, _ := ParseFloat([]byte(tc.input))
			if got != tc.expected && !tc.expectedPanic {
				t.Errorf("ParseFloat(%s): expected %v, got %v",
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
			got, _, _ := ParseFloat([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("ParseFloat(%s): got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

type ParseIntTest struct {
	in         string
	out        int64
	isErr      bool
	isOverflow bool
}

var parseIntTests = []ParseIntTest{
	{
		in:  "0",
		out: 0,
	},
	{
		in:  "1",
		out: 1,
	},
	{
		in:  "-1",
		out: -1,
	},
	{
		in:  "12345",
		out: 12345,
	},
	{
		in:  "-12345",
		out: -12345,
	},
	{
		in:  "9223372036854775807", // = math.MaxInt64
		out: 9223372036854775807,
	},
	{
		in:  "-9223372036854775808", // = math.MinInt64
		out: -9223372036854775808,
	},
	{
		in:         "-92233720368547758081",
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
	{
		in:         "18446744073709551616", // = 2^64
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
	{
		in:         "9223372036854775808", // = math.MaxInt64 - 1
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
	{
		in:         "-9223372036854775809", // = math.MaxInt64 - 1
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
	{
		in:    "",
		isErr: true,
	},
	{
		in:    "abc",
		isErr: true,
	},
	{
		in:    "12345x",
		isErr: true,
	},
	{
		in:    "123e5",
		isErr: true,
	},
	{
		in:    "9223372036854775807x",
		isErr: true,
	},
	{
		in:         "27670116110564327410",
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
	{
		in:         "-27670116110564327410",
		out:        0,
		isErr:      true,
		isOverflow: true,
	},
}

func TestBytesParseInt(t *testing.T) {
	for _, test := range parseIntTests {
		out, ok, overflow := ParseInt([]byte(test.in))
		if overflow != test.isOverflow {
			t.Errorf("Test '%s' error return did not overflow expectation (obtained %t, expected %t)", test.in, overflow, test.isOverflow)
		}
		if ok != !test.isErr {
			t.Errorf("Test '%s' error return did not match expectation (obtained %t, expected %t)", test.in, !ok, test.isErr)
		} else if ok && out != test.out {
			t.Errorf("Test '%s' did not return the expected value (obtained %d, expected %d)", test.in, out, test.out)
		}
	}
}