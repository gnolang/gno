package uint256

import (
	"strings"
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
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
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
		x, err := FromDecimal(tc.x)
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

func TestLtUint64(t *testing.T) {
	tests := []struct {
		x    string
		y    uint64
		want bool
	}{
		{"0", 1, true},
		{"1", 0, false},
		{"10", 10, false},
		{"0xffffffffffffffff", 0, false},
		{"0x10000000000000000", 10000000000000000, false},
	}

	for _, tc := range tests {
		var x *Uint
		var err error

		if strings.HasPrefix(tc.x, "0x") {
			x, err = FromHex(tc.x)
			if err != nil {
				t.Error(err)
				continue
			}
		} else {
			x, err = FromDecimal(tc.x)
			if err != nil {
				t.Error(err)
				continue
			}
		}

		got := x.LtUint64(tc.y)

		if got != tc.want {
			t.Errorf("LtUint64(%s, %d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestSGT(t *testing.T) {
	x := MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	y := MustFromHex("0x0")
	actual := x.Sgt(y)
	if actual {
		t.Fatalf("Expected %v false", actual)
	}

	x = MustFromHex("0x0")
	y = MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	actual = x.Sgt(y)
	if !actual {
		t.Fatalf("Expected %v true", actual)
	}
}

func TestEq(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{"0xffffffffffffffff", "18446744073709551615", true},
		{"0x10000000000000000", "18446744073709551616", true},
		{"0", "0", true},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639935", true},
	}

	for i, tc := range tests {
		var x *Uint
		var err error

		if strings.HasPrefix(tc.x, "0x") {
			x, err = FromHex(tc.x)
			if err != nil {
				t.Error(err)
				continue
			}
		} else {
			x, err = FromDecimal(tc.x)
			if err != nil {
				t.Error(err)
				continue
			}
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.Eq(y)

		if got != tc.want {
			t.Errorf("Eq(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}
