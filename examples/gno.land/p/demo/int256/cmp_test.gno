package int256

import (
	"testing"
)

func TestEq(t *testing.T) {
	tests := []struct {
		x, y string
		want bool
	}{
		{"0", "0", true},
		{"0", "1", false},
		{"1", "0", false},
		{"-1", "0", false},
		{"0", "-1", false},
		{"1", "1", true},
		{"-1", "-1", true},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", false},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", true},
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

		got := x.Eq(y)
		if got != tc.want {
			t.Errorf("Eq(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestNeq(t *testing.T) {
	tests := []struct {
		x, y string
		want bool
	}{
		{"0", "0", false},
		{"0", "1", true},
		{"1", "0", true},
		{"-1", "0", true},
		{"0", "-1", true},
		{"1", "1", false},
		{"-1", "-1", false},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", true},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", false},
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

		got := x.Neq(y)
		if got != tc.want {
			t.Errorf("Neq(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		x, y string
		want int
	}{
		{"0", "0", 0},
		{"0", "1", -1},
		{"1", "0", 1},
		{"-1", "0", -1},
		{"0", "-1", 1},
		{"1", "1", 0},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", 1},
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
		{"-0", true},
		{"1", false},
		{"-1", false},
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

func TestIsNeg(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0", false},
		{"-0", true}, // TODO: should this be false?
		{"1", false},
		{"-1", true},
		{"10", false},
		{"-10", true},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.IsNeg()
		if got != tc.want {
			t.Errorf("IsNeg(%s) = %v, want %v", tc.x, got, tc.want)
		}
	}
}

func TestLt(t *testing.T) {
	tests := []struct {
		x, y string
		want bool
	}{
		{"0", "0", false},
		{"0", "1", true},
		{"1", "0", false},
		{"-1", "0", true},
		{"0", "-1", false},
		{"1", "1", false},
		{"-1", "-1", false},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", false},
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

		got := x.Lt(y)
		if got != tc.want {
			t.Errorf("Lt(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestGt(t *testing.T) {
	tests := []struct {
		x, y string
		want bool
	}{
		{"0", "0", false},
		{"0", "1", false},
		{"1", "0", true},
		{"-1", "0", false},
		{"0", "-1", true},
		{"1", "1", false},
		{"-1", "-1", false},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", true},
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

		got := x.Gt(y)
		if got != tc.want {
			t.Errorf("Gt(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestClone(t *testing.T) {
	tests := []struct {
		x string
	}{
		{"0"},
		{"-0"},
		{"1"},
		{"-1"},
		{"10"},
		{"-10"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y := x.Clone()

		if x.Cmp(y) != 0 {
			t.Errorf("Clone(%s) = %v, want %v", tc.x, y, x)
		}
	}
}
