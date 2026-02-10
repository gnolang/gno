package int256

import (
	"testing"

	"gno.land/p/onbloc/uint256"
)

func TestInitializers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() *Int
		wantSign int
		wantStr  string
	}{
		{"Zero", Zero, 0, "0"},
		{"New", New, 0, "0"},
		{"One", One, 1, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := tt.fn()
			if z.Sign() != tt.wantSign {
				t.Errorf("%s() = %d, want %d", tt.name, z.Sign(), tt.wantSign)
			}
			if z.String() != tt.wantStr {
				t.Errorf("%s() = %s, want %s", tt.name, z.String(), tt.wantStr)
			}
		})
	}
}

func TestNewInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected int
	}{
		{0, 0},
		{1, 1},
		{-1, -1},
		{9223372036854775807, 1},   // max int64
		{-9223372036854775808, -1}, // min int64
	}

	for _, tt := range tests {
		z := NewInt(tt.input)
		if z.Sign() != tt.expected {
			t.Errorf("NewInt(%d) = %d, want %d", tt.input, z.Sign(), tt.expected)
		}
	}
}

func TestFromDecimal(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		z, err := FromDecimal(tt.input)
		if tt.isError {
			if err == nil {
				t.Errorf("FromDecimal(%s) expected error, but got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("FromDecimal(%s) unexpected error: %v", tt.input, err)
			} else if z.Sign() != tt.expected {
				t.Errorf("FromDecimal(%s) sign is incorrect. Expected: %d, Actual: %d", tt.input, tt.expected, z.Sign())
			}
		}
	}
}

func TestMustFromDecimal(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		shouldPanic bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"123", 1, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustFromDecimal(%q) expected panic, but got nil", tt.input)
				}
			}()
		}

		z := MustFromDecimal(tt.input)
		if !tt.shouldPanic && z.Sign() != tt.expected {
			t.Errorf("MustFromDecimal(%q) sign is incorrect. Expected: %d, Actual: %d", tt.input, tt.expected, z.Sign())
		}
	}
}

func TestSetUint64(t *testing.T) {
	tests := []uint64{
		0,
		1,
		18446744073709551615, // max uint64
	}

	for _, tt := range tests {
		z := New().SetUint64(tt)
		if z.Sign() < 0 {
			t.Errorf("SetUint64(%d) result is negative", tt)
		}
		if tt == 0 && z.Sign() != 0 {
			t.Errorf("SetUint64(0) result is not zero")
		}
		if tt > 0 && z.Sign() != 1 {
			t.Errorf("SetUint64(%d) result is not positive", tt)
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

	for _, tt := range tests {
		z := New().FromUint256(tt.input)
		if z.Sign() != tt.expected {
			t.Errorf("FromUint256(%v) = %d, want %d", tt.input, z.Sign(), tt.expected)
		}
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		x    string
		want int
	}{
		{"0", 0},
		{"-0", 0},
		{"+0", 0},
		{"1", 1},
		{"-1", -1},
		{"9223372036854775807", 1},
		{"-9223372036854775808", -1},
	}

	for _, tt := range tests {
		z := MustFromDecimal(tt.x)
		got := z.Sign()
		if got != tt.want {
			t.Errorf("Sign(%s) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func BenchmarkSign(b *testing.B) {
	z := New()
	for i := 0; i < b.N; i++ {
		z.SetUint64(uint64(i))
		z.Sign()
	}
}

func TestSetAndToString(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		z, err := New().SetString(tt.input)
		if tt.isError {
			if err == nil {
				t.Errorf("SetString(%s) expected error, but got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("SetString(%s) unexpected error: %v", tt.input, err)
			} else if z.Sign() != tt.expected {
				t.Errorf("SetString(%s) sign is incorrect. Expected: %d, Actual: %d", tt.input, tt.expected, z.Sign())
			} else if z.String() != tt.input {
				t.Errorf("SetString(%s) string representation is incorrect. Expected: %s, Actual: %s", tt.input, tt.input, z.String())
			}
		}
	}
}
