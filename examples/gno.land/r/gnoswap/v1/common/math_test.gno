package common

import (
	"testing"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestI64Min(t *testing.T) {
	tests := []struct {
		x, y, want int64
	}{
		{1, 2, 1},
		{-1, 1, -1},
		{5, 5, 5},
		{-10, -5, -10},
		{9223372036854775807, 0, 0}, // Max int64
	}

	for _, tt := range tests {
		got := I64Min(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("I64Min(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestI64Max(t *testing.T) {
	tests := []struct {
		x, y, want int64
	}{
		{1, 2, 2},
		{-1, 1, 1},
		{5, 5, 5},
		{-10, -5, -5},
		{9223372036854775807, 0, 9223372036854775807}, // Max int64
	}

	for _, tt := range tests {
		got := I64Max(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("I64Max(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestU64Min(t *testing.T) {
	tests := []struct {
		x, y, want uint64
	}{
		{1, 2, 1},
		{0, 1, 0},
		{5, 5, 5},
		{10, 5, 5},
		{18446744073709551615, 0, 0}, // Max uint64
	}

	for _, tt := range tests {
		got := U64Min(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("U64Min(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestU64Max(t *testing.T) {
	tests := []struct {
		x, y, want uint64
	}{
		{1, 2, 2},
		{0, 1, 1},
		{5, 5, 5},
		{10, 5, 10},
		{18446744073709551615, 0, 18446744073709551615}, // Max uint64
	}

	for _, tt := range tests {
		got := U64Max(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("U64Max(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestI256Min(t *testing.T) {
	tests := []struct {
		x, y string // hex strings for creating Int
		want string
	}{
		{"1", "2", "1"},
		{"-1", "1", "-1"},
		{"5", "5", "5"},
		{"-10", "-5", "-10"},
	}

	for _, tt := range tests {
		x, _ := i256.FromDecimal(tt.x)
		y, _ := i256.FromDecimal(tt.y)
		want, _ := i256.FromDecimal(tt.want)
		got := I256Min(x, y)
		if got.Cmp(want) != 0 {
			t.Errorf("I256Min(%v, %v) = %v, want %v", tt.x, tt.y, got, want)
		}
	}
}

func TestI256Max(t *testing.T) {
	tests := []struct {
		x, y string // hex strings for creating Int
		want string
	}{
		{"1", "2", "2"},
		{"-1", "1", "1"},
		{"5", "5", "5"},
		{"-10", "-5", "-5"},
	}

	for _, tt := range tests {
		x, _ := i256.FromDecimal(tt.x)
		y, _ := i256.FromDecimal(tt.y)
		want, _ := i256.FromDecimal(tt.want)
		got := I256Max(x, y)
		if got.Cmp(want) != 0 {
			t.Errorf("I256Max(%v, %v) = %v, want %v", tt.x, tt.y, got, want)
		}
	}
}

func TestU256Min(t *testing.T) {
	tests := []struct {
		x, y string // decimal strings for creating Uint
		want string
	}{
		{"1", "2", "1"},
		{"0", "1", "0"},
		{"5", "5", "5"},
		{"10", "5", "5"},
	}

	for _, tt := range tests {
		x, _ := u256.FromDecimal(tt.x)
		y, _ := u256.FromDecimal(tt.y)
		want, _ := u256.FromDecimal(tt.want)
		got := U256Min(x, y)
		if got.Cmp(want) != 0 {
			t.Errorf("U256Min(%v, %v) = %v, want %v", tt.x, tt.y, got, want)
		}
	}
}

func TestU256Max(t *testing.T) {
	tests := []struct {
		x, y string // decimal strings for creating Uint
		want string
	}{
		{"1", "2", "2"},
		{"0", "1", "1"},
		{"5", "5", "5"},
		{"10", "5", "10"},
	}

	for _, tt := range tests {
		x, _ := u256.FromDecimal(tt.x)
		y, _ := u256.FromDecimal(tt.y)
		want, _ := u256.FromDecimal(tt.want)
		got := U256Max(x, y)
		if got.Cmp(want) != 0 {
			t.Errorf("U256Max(%v, %v) = %v, want %v", tt.x, tt.y, got, want)
		}
	}
}

func TestSafeConvertUint256ToInt256(t *testing.T) {
	tests := []struct {
		x, want     string
		shouldPanic bool
	}{
		{"0", "0", false},
		{"1", "1", false},
		// max int256
		{"57896044618658097711785492504343953926634992332820282019728792003956564819968", "57896044618658097711785492504343953926634992332820282019728792003956564819968", false},
		// max int256 + 1 (overflow)
		{"57896044618658097711785492504343953926634992332820282019728792003956564819969", "", true},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("SafeConvertUint256ToInt256(%v) did not panic", tt.x)
				}
			}()
		}

		x := u256.MustFromDecimal(tt.x)
		got := SafeConvertUint256ToInt256(x)
		want := i256.MustFromDecimal(tt.want)
		if got.Cmp(want) != 0 {
			t.Errorf("SafeConvertUint256ToInt256(%v) = %v, want %v", tt.x, got, want)
		}
	}
}

func TestSafeConvertUint256ToUint64(t *testing.T) {
	tests := []struct {
		x           string
		want        uint64
		shouldPanic bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		// max uint64
		{"18446744073709551615", 18446744073709551615, false},
		// max uint64 + 1 (overflow)
		{"18446744073709551616", 0, true},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("SafeConvertUint256ToUint64(%v) did not panic", tt.x)
				}
			}()
		}

		x := u256.MustFromDecimal(tt.x)
		got := SafeConvertUint256ToUint64(x)
		if got != tt.want {
			t.Errorf("SafeConvertUint256ToUint64(%v) = %v, want %v", tt.x, got, tt.want)
		}
	}
}
