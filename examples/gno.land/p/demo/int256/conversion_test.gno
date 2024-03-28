package int256

import (
	"testing"
)

func TestSetInt64(t *testing.T) {
	tests := []struct {
		x    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{9223372036854775807, "9223372036854775807"},
		{-9223372036854775808, "-9223372036854775808"},

		// bigint overflows target kind
		// {9223372036854775808, "9223372036854775808"},

		// bigint underflows target kind:
		// {-9223372036854775809, "-9223372036854775809"},
	}

	for _, tc := range tests {
		var z Int
		z.SetInt64(tc.x)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("SetInt64(%d) = %s, want %s", tc.x, got, tc.want)
		}
	}
}

func TestSetUint64(t *testing.T) {
	tests := []struct {
		x    uint64
		want string
	}{
		{0, "0"},
		{1, "1"},
		// bigint overflows target kind
		//{115792089237316195423570985008687907853269984665640564039457584007913129639935, "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range tests {
		var z Int
		z.SetUint64(tc.x)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("SetUint64(%d) = %s, want %s", tc.x, got, tc.want)
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
		{"18446744073709551616", 0},
		{"18446744073709551617", 1},
		{"-1", 1},
		{"-18446744073709551615", 18446744073709551615},
		{"-18446744073709551616", 0},
		{"-18446744073709551617", 1},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)

		got := z.Uint64()
		if got != tc.want {
			t.Errorf("Uint64(%s) = %d, want %d", tc.x, got, tc.want)
		}
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
		//{"9223372036854775808", 9223372036854775808}, // overflow int64
		{"18446744073709551616", 0},
		{"18446744073709551617", 1},
		{"-1", -1},
		{"-9223372036854775808", -9223372036854775808},
		//{"-9223372036854775809", -9223372036854775809}, // underflow int64
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)

		got := z.Int64()
		if got != tc.want {
			t.Errorf("Uint64(%s) = %d, want %d", tc.x, got, tc.want)
		}
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

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)
		z.Neg(z)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("Neg(%s) = %s, want %s", tc.x, got, tc.want)
		}
	}
}
