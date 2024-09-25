package int256

import (
	"testing"

	"gno.land/p/demo/uint256"
)

func TestZero(t *testing.T) {
	z := Zero()
	if z.Sign() != 0 {
		t.Errorf("Zero() = %d, want %d", z.Sign(), 0)
	}
}

func TestNew(t *testing.T) {
	z := New()
	if z.Sign() != 0 {
		t.Errorf("New() = %d, want %d", z.Sign(), 0)
	}
}

func TestNewInt(t *testing.T) {
	testCases := []struct {
		input    int64
		expected int
	}{
		{0, 0},
		{1, 1},
		{-1, -1},
		{9223372036854775807, 1},   // max int64
		{-9223372036854775808, -1}, // min int64
	}

	for _, tc := range testCases {
		z := NewInt(tc.input)
		if z.Sign() != tc.expected {
			t.Errorf("NewInt(%d) = %d, want %d", tc.input, z.Sign(), tc.expected)
		}
	}
}

func TestFromDecimal(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
		isError  bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"123456789", 1, false},
		{"-123456789", -1, false},
		{"invalid", 0, true},
	}

	for _, tc := range testCases {
		z, err := FromDecimal(tc.input)
		if tc.isError {
			if err == nil {
				t.Errorf("FromDecimal(%s) expected error, but got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("FromDecimal(%s) unexpected error: %v", tc.input, err)
			} else if z.Sign() != tc.expected {
				t.Errorf("FromDecimal(%s) sign is incorrect. Expected: %d, Actual: %d", tc.input, tc.expected, z.Sign())
			}
		}
	}
}

func TestMustFromDecimal(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustFromDecimal(\"invalid\") expected panic, but got nil")
		}
	}()

	z := MustFromDecimal("123")
	if z.Sign() != 1 {
		t.Errorf("MustFromDecimal(\"123\") sign is incorrect. Expected: %d, Actual: %d", 1, z.Sign())
	}

	MustFromDecimal("invalid")
}

func TestSetUint64(t *testing.T) {
	testCases := []uint64{
		0,
		1,
		18446744073709551615, // max uint64
	}

	for _, tc := range testCases {
		z := New().SetUint64(tc)
		if z.Sign() < 0 {
			t.Errorf("SetUint64(%d) result is negative", tc)
		}
		if tc == 0 && z.Sign() != 0 {
			t.Errorf("SetUint64(0) result is not zero")
		}
		if tc > 0 && z.Sign() != 1 {
			t.Errorf("SetUint64(%d) result is not positive", tc)
		}
	}
}

func TestFromUint256(t *testing.T) {
	tests := []struct {
		input    *uint256.Uint
		expected int
	}{
		{uint256.NewUint(0), 0},
		{uint256.NewUint(1), 1},
		{uint256.NewUint(18446744073709551615), 1},
	}

	for _, tc := range tests {
		z := New().FromUint256(tc.input)
		if z.Sign() != tc.expected {
			t.Errorf("FromUint256(%v) = %d, want %d", tc.input, z.Sign(), tc.expected)
		}
	}
}

func TestNilToZero(t *testing.T) {
	z := New().NilToZero()
	if z.Sign() != 0 {
		t.Errorf("NilToZero() = %d, want %d", z.Sign(), 0)
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		x    string
		want int
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)
		got := z.Sign()
		if got != tc.want {
			t.Errorf("Sign(%s) = %d, want %d", tc.x, got, tc.want)
		}
	}
}
