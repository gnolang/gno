package int256

import (
	"testing"

	"gno.land/p/onbloc/uint256"
)

func TestSetInt64(t *testing.T) {
	tests := []struct {
		v      int64
		expect int
	}{
		{0, 0},
		{1, 1},
		{-1, -1},
		{9223372036854775807, 1},   // overflow (max int64)
		{-9223372036854775808, -1}, // underflow (min int64)
	}

	for _, tt := range tests {
		z := New().SetInt64(tt.v)
		if z.Sign() != tt.expect {
			t.Errorf("SetInt64(%d) = %d, want %d", tt.v, z.Sign(), tt.expect)
		}
	}
}

func TestUint64(t *testing.T) {
	tests := []struct {
		x    string
		want uint64
	}{
		{"0", 0},
		{"1", 1},
		{"9223372036854775807", 9223372036854775807},
		{"9223372036854775808", 9223372036854775808},
		{"18446744073709551615", 18446744073709551615},
	}

	for _, tt := range tests {
		z := MustFromDecimal(tt.x)

		got := z.Uint64()
		if got != tt.want {
			t.Errorf("Uint64(%s) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func TestUint64_Panic(t *testing.T) {
	tests := []struct {
		x string
	}{
		{"-1"},
		{"18446744073709551616"},
		{"18446744073709551617"},
	}

	for _, tt := range tests {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Uint64(%s) did not panic", tt.x)
			}
		}()

		z := MustFromDecimal(tt.x)
		z.Uint64()
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		x    string
		want int64
	}{
		{"0", 0},
		{"1", 1},
		{"9223372036854775807", 9223372036854775807},
		{"-1", -1},
		{"-9223372036854775808", -9223372036854775808},
	}

	for _, tt := range tests {
		z := MustFromDecimal(tt.x)

		got := z.Int64()
		if got != tt.want {
			t.Errorf("Uint64(%s) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func TestInt64_Panic(t *testing.T) {
	tests := []struct {
		x string
	}{
		{"18446744073709551616"},
		{"18446744073709551617"},
	}

	for _, tt := range tests {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Int64(%s) did not panic", tt.x)
			}
		}()

		z := MustFromDecimal(tt.x)
		z.Int64()
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "-1"},
		{"-1", "1"},
		{"9223372036854775807", "-9223372036854775807"},
		{"-18446744073709551615", "18446744073709551615"},
	}

	for _, tt := range tests {
		z := MustFromDecimal(tt.x)
		z.Neg(z)

		got := z.String()
		if got != tt.want {
			t.Errorf("Neg(%s) = %s, want %s", tt.x, got, tt.want)
		}
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"9223372036854775807", "9223372036854775807"},
		{"-18446744073709551615", "-18446744073709551615"},
	}

	for _, tt := range tests {
		z := MustFromDecimal(tt.x)
		z.Set(z)

		got := z.String()
		if got != tt.want {
			t.Errorf("Set(%s) = %s, want %s", tt.x, got, tt.want)
		}
	}
}

func TestSetUint256(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"9223372036854775807", "9223372036854775807"},
		{"18446744073709551615", "18446744073709551615"},
	}

	for _, tt := range tests {
		got := New()

		z := uint256.MustFromDecimal(tt.x)
		got.SetUint256(z)

		if got.String() != tt.want {
			t.Errorf("SetUint256(%s) = %s, want %s", tt.x, got.String(), tt.want)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"123456789", "123456789"},
		{"-123456789", "-123456789"},
		{"18446744073709551615", "18446744073709551615"}, // max uint64
		{"-18446744073709551615", "-18446744073709551615"},
		{TWO_POW_128_MINUS_1, TWO_POW_128_MINUS_1},
		{MINUS_TWO_POW_128, MINUS_TWO_POW_128},
		{MIN_INT256, MIN_INT256},
		{MAX_INT256, MAX_INT256},
	}

	for _, tt := range tests {
		x, err := FromDecimal(tt.input)
		if err != nil {
			t.Errorf("Failed to parse input (%s): %v", tt.input, err)
			continue
		}

		output := x.String()

		if output != tt.expected {
			t.Errorf("String(%s) = %s, want %s", tt.input, output, tt.expected)
		}
	}
}

func TestNilToZero(t *testing.T) {
	z := New().NilToZero()
	if z.Sign() != 0 {
		t.Errorf("NilToZero() = %d, want %d", z.Sign(), 0)
	}
}
