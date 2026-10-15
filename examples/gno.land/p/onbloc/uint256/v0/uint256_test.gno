package uint256

import (
	"testing"
)

func TestSetAllOne(t *testing.T) {
	z := Zero()
	z.SetAllOne()
	if z.String() != twoPow256Sub1 {
		t.Errorf("Expected all ones, got %s", z.String())
	}
}

func TestByte(t *testing.T) {
	tests := []struct {
		input    string
		position uint64
		expected byte
	}{
		{"0x1000000000000000000000000000000000000000000000000000000000000000", 0, 16},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, 255},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 31, 255},
	}

	for i, tt := range tests {
		z, _ := FromHex(tt.input)
		n := NewUint(tt.position)
		result := z.Byte(n)

		if result.arr[0] != uint64(tt.expected) {
			t.Errorf("Test case %d failed. Input: %s, Position: %d, Expected: %d, Got: %d",
				i, tt.input, tt.position, tt.expected, result.arr[0])
		}

		// check other array elements are 0
		if result.arr[1] != 0 || result.arr[2] != 0 || result.arr[3] != 0 {
			t.Errorf("Test case %d failed. Non-zero values in upper bytes", i)
		}
	}

	// overflow
	z, _ := FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	n := NewUint(32)
	result := z.Byte(n)

	if !result.IsZero() {
		t.Errorf("Expected zero for position >= 32, got %v", result)
	}
}

func TestBitLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0x0", 0},
		{"0x1", 1},
		{"0xff", 8},
		{"0x100", 9},
		{"0xffff", 16},
		{"0x10000", 17},
		{"0xffffffffffffffff", 64},
		{"0x10000000000000000", 65},
		{"0xffffffffffffffffffffffffffffffff", 128},
		{"0x100000000000000000000000000000000", 129},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 256},
	}

	for i, tt := range tests {
		z, _ := FromHex(tt.input)
		result := z.BitLen()

		if result != tt.expected {
			t.Errorf("Test case %d failed. Input: %s, Expected: %d, Got: %d",
				i, tt.input, tt.expected, result)
		}
	}
}

func TestByteLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0x0", 0},
		{"0x1", 1},
		{"0xff", 1},
		{"0x100", 2},
		{"0xffff", 2},
		{"0x10000", 3},
		{"0xffffffffffffffff", 8},
		{"0x10000000000000000", 9},
		{"0xffffffffffffffffffffffffffffffff", 16},
		{"0x100000000000000000000000000000000", 17},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 32},
	}

	for i, tt := range tests {
		z, _ := FromHex(tt.input)
		result := z.ByteLen()

		if result != tt.expected {
			t.Errorf("Test case %d failed. Input: %s, Expected: %d, Got: %d",
				i, tt.input, tt.expected, result)
		}
	}
}

func TestClone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0x1", "1"},
		{"0x100", "256"},
		{"0x10000000000000000", "18446744073709551616"},
	}

	for _, tt := range tests {
		z, _ := FromHex(tt.input)
		result := z.Clone()
		if result.String() != tt.expected {
			t.Errorf("Test %s failed. Expected %s, got %s", tt.input, tt.expected, result.String())
		}
	}
}
