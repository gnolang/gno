package int256

import (
	"testing"

	"gno.land/p/demo/uint256"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		x, want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"-1", "1"},
		{"-2", "2"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.Abs()

		if got.ToString() != tc.want {
			t.Errorf("Abs(%s) = %v, want %v", tc.x, got.ToString(), tc.want)
		}
	}
}

func TestAbsGt(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "0", "false"},
		{"1", "0", "true"},
		{"-1", "0", "true"},
		{"-1", "1", "false"},
		{"-2", "1", "true"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0", "true"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "1", "true"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639935", "false"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := uint256.FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.AbsGt(y)

		if got != (tc.want == "true") {
			t.Errorf("AbsGt(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestAbsLt(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "0", "false"},
		{"1", "0", "false"},
		{"-1", "0", "false"},
		{"-1", "1", "false"},
		{"-2", "1", "false"},
		{"-5", "10", "true"},
		{"31330", "31337", "true"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0", "false"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "1", "false"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639935", "false"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := uint256.FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.AbsLt(y)

		if got != (tc.want == "true") {
			t.Errorf("AbsLt(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}
