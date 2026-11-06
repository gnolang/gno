package int256

import (
	"testing"
)

func TestBitwise_And(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"5", "1", "1"},  // 0101 & 0001 = 0001
		{"-1", "1", "1"}, // 1111 & 0001 = 0001
		{"-5", "3", "3"}, // 1111...1011 & 0000...0011 = 0000...0011
		{MAX_UINT256, MAX_UINT256, MAX_UINT256},
		{TWO_POW_128, TWO_POW_128_MINUS_1, "0"}, // 2^128 & (2^128 - 1) = 0
		{TWO_POW_128, MAX_UINT256, TWO_POW_128}, // 2^128 & MAX_INT256
		{MAX_UINT256, TWO_POW_128, TWO_POW_128}, // MAX_INT256 & 2^128
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		y, _ := FromDecimal(tc.y)
		want, _ := FromDecimal(tc.want)

		got := new(Int).And(x, y)

		if got.Neq(want) {
			t.Errorf("And(%s, %s) = %s, want %s", x.String(), y.String(), got.String(), want.String())
		}
	}
}

func TestBitwise_Or(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"5", "1", "5"},   // 0101 | 0001 = 0101
		{"-1", "1", "-1"}, // 1111 | 0001 = 1111
		{"-5", "3", "-5"}, // 1111...1011 | 0000...0011 = 1111...1011
		{TWO_POW_128, TWO_POW_128_MINUS_1, TWO_POW_129_MINUS_1},
		{TWO_POW_128, MAX_UINT256, MAX_UINT256},
		{"0", TWO_POW_128, TWO_POW_128},         // 0 | 2^128 = 2^128
		{MAX_UINT256, TWO_POW_128, MAX_UINT256}, // MAX_INT256 | 2^128 = MAX_INT256
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		y, _ := FromDecimal(tc.y)
		want, _ := FromDecimal(tc.want)

		got := new(Int).Or(x, y)

		if got.Neq(want) {
			t.Errorf(
				"Or(%s, %s) = %s, want %s",
				x.String(), y.String(), got.String(), want.String(),
			)
		}
	}
}

func TestBitwise_Not(t *testing.T) {
	tests := []struct {
		x, want string
	}{
		{"5", "-6"},                              // 0101 -> 1111...1010
		{"-1", "0"},                              // 1111...1111 -> 0000...0000
		{TWO_POW_128, MINUS_TWO_POW_128_MINUS_1}, // NOT 2^128
		{TWO_POW_255, MIN_INT256_MINUS_1},        // NOT 2^255
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		want, _ := FromDecimal(tc.want)

		got := new(Int).Not(x)

		if got.Neq(want) {
			t.Errorf("Not(%s) = %s, want %s", x.String(), got.String(), want.String())
		}
	}
}

func TestBitwise_Xor(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"5", "1", "4"},                 // 0101 ^ 0001 = 0100
		{"-1", "1", "-2"},               // 1111...1111 ^ 0000...0001 = 1111...1110
		{"-5", "3", "-8"},               // 1111...1011 ^ 0000...0011 = 1111...1000
		{TWO_POW_128, TWO_POW_128, "0"}, // 2^128 ^ 2^128 = 0
		{MAX_UINT256, TWO_POW_128, MINUS_TWO_POW_128_MINUS_1}, // MAX_INT256 ^ 2^128
		{TWO_POW_255, MAX_UINT256, MIN_INT256_MINUS_1},        // 2^255 ^ MAX_INT256
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		y, _ := FromDecimal(tt.y)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Xor(x, y)

		if got.Neq(want) {
			t.Errorf("Xor(%s, %s) = %s, want %s", x.String(), y.String(), got.String(), want.String())
		}
	}
}

func TestBitwise_Rsh(t *testing.T) {
	tests := []struct {
		x    string
		n    uint
		want string
	}{
		{"5", 1, "2"},  // 0101 >> 1 = 0010
		{"42", 3, "5"}, // 00101010 >> 3 = 00000101
		{TWO_POW_128, 128, "1"},
		{MAX_UINT256, 255, "1"},
		{TWO_POW_255, 254, "2"},
		{MINUS_TWO_POW_128, 128, TWO_POW_128_MINUS_1},
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Rsh(x, tt.n)

		if got.Neq(want) {
			t.Errorf("Rsh(%s, %d) = %s, want %s", x.String(), tt.n, got.String(), want.String())
		}
	}
}

func TestBitwise_Lsh(t *testing.T) {
	tests := []struct {
		x    string
		n    uint
		want string
	}{
		{"5", 2, "20"},          // 0101 << 2 = 10100
		{"42", 5, "1344"},       // 00101010 << 5 = 10101000000
		{"1", 128, TWO_POW_128}, // 1 << 128 = 2^128
		{"2", 254, TWO_POW_255},
		{"1", 255, MIN_INT256}, // 1 << 255 = MIN_INT256 (overflow)
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Lsh(x, tt.n)

		if got.Neq(want) {
			t.Errorf("Lsh(%s, %d) = %s, want %s", x.String(), tt.n, got.String(), want.String())
		}
	}
}
