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
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		y, _ := FromDecimal(tc.y)
		want, _ := FromDecimal(tc.want)

		got := new(Int).And(x, y)

		if got.Neq(want) {
			t.Errorf("And(%s, %s) = %s, want %s", x.ToString(), y.ToString(), got.ToString(), want.ToString())
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
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		y, _ := FromDecimal(tc.y)
		want, _ := FromDecimal(tc.want)

		got := new(Int).Or(x, y)

		if got.Neq(want) {
			t.Errorf("Or(%s, %s) = %s, want %s", x.ToString(), y.ToString(), got.ToString(), want.ToString())
		}
	}
}

func TestBitwise_Not(t *testing.T) {
	tests := []struct {
		x, want string
	}{
		{"5", "-6"}, // 0101 -> 1111...1010
		{"-1", "0"}, // 1111...1111 -> 0000...0000
	}

	for _, tc := range tests {
		x, _ := FromDecimal(tc.x)
		want, _ := FromDecimal(tc.want)

		got := new(Int).Not(x)

		if got.Neq(want) {
			t.Errorf("Not(%s) = %s, want %s", x.ToString(), got.ToString(), want.ToString())
		}
	}
}

func TestBitwise_Xor(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"5", "1", "4"},   // 0101 ^ 0001 = 0100
		{"-1", "1", "-2"}, // 1111...1111 ^ 0000...0001 = 1111...1110
		{"-5", "3", "-8"}, // 1111...1011 ^ 0000...0011 = 1111...1000
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		y, _ := FromDecimal(tt.y)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Xor(x, y)

		if got.Neq(want) {
			t.Errorf("Xor(%s, %s) = %s, want %s", x.ToString(), y.ToString(), got.ToString(), want.ToString())
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
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Rsh(x, tt.n)

		if got.Neq(want) {
			t.Errorf("Rsh(%s, %d) = %s, want %s", x.ToString(), tt.n, got.ToString(), want.ToString())
		}
	}
}

func TestBitwise_Lsh(t *testing.T) {
	tests := []struct {
		x    string
		n    uint
		want string
	}{
		{"5", 2, "20"},    // 0101 << 2 = 10100
		{"42", 5, "1344"}, // 00101010 << 5 = 10101000000
	}

	for _, tt := range tests {
		x, _ := FromDecimal(tt.x)
		want, _ := FromDecimal(tt.want)

		got := new(Int).Lsh(x, tt.n)

		if got.Neq(want) {
			t.Errorf("Lsh(%s, %d) = %s, want %s", x.ToString(), tt.n, got.ToString(), want.ToString())
		}
	}
}
