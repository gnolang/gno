package ryu

import (
	"math"
	"testing"
)

func TestFormatFloat64(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"positive infinity", math.Inf(1), "+Inf"},
		{"negative infinity", math.Inf(-1), "-Inf"},
		{"NaN", math.NaN(), "NaN"},
		{"zero", 0.0, "0e+00"},
		{"negative zero", -0.0, "0e+00"},
		{"positive number", 3.14159, "3.14159e+00"},
		{"negative number", -2.71828, "-2.71828e+00"},
		{"very small number", 1.23e-20, "1.23e-20"},
		{"very large number", 1.23e+20, "1.23e+20"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatFloat64(test.value)
			if result != test.expected {
				t.Errorf("FormatFloat64(%v) = %q, expected %q", test.value, result, test.expected)
			}
		})
	}
}
