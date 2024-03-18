package uint256

import (
	"testing"
)

func TestCmp(t *testing.T) {
	tests := []struct {
		x, y string
		want int
	}{
		{"0", "0", 0},
		{"0", "1", -1},
		{"1", "0", 1},
		{"1", "1", 0},
		{"10", "10", 0},
		{"10", "11", -1},
		{"11", "10", 1},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.Cmp(y)
		if got != tc.want {
			t.Errorf("Cmp(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestLtUint64(t *testing.T) {
	tests := []struct {
		x    string
		y    uint64
		want bool
	}{
		{"0", 1, true},
		{"1", 0, false},
		{"10", 10, false},
		{"ffffffffffffffff", 0, false},
		{"10000000000000000", 0, false},
	}

	for _, tc := range tests {
		x := new(Uint).SetBytes(hex2Bytes(tc.x))
		got := x.LtUint64(tc.y)

		if got != tc.want {
			t.Errorf("LtUint64(%s, %d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0", true},
		{"1", false},
		{"10", false},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.IsZero()
		if got != tc.want {
			t.Errorf("IsZero(%s) = %v, want %v", tc.x, got, tc.want)
		}
	}
}

func TestSGT(t *testing.T) {
	x := new(Uint).SetBytes(hex2Bytes("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe"))
	y := new(Uint).SetBytes(hex2Bytes("00"))
	actual := x.Sgt(y)
	if actual {
		t.Fatalf("Expected %v false", actual)
	}

	x = new(Uint).SetBytes(hex2Bytes("00"))
	y = new(Uint).SetBytes(hex2Bytes("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe"))
	actual = x.Sgt(y)
	if !actual {
		t.Fatalf("Expected %v true", actual)
	}
}
