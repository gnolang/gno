package int256

import (
	"testing"

	"gno.land/p/gnoswap/uint256"
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

func TestInt_AbsOverflow(t *testing.T) {
	tests := []struct {
		name         string
		x            *Int
		wantResult   string
		wantOverflow bool
	}{
		{
			name:         "zero",
			x:            Zero(),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "positive number",
			x:            NewInt(100),
			wantResult:   "100",
			wantOverflow: false,
		},
		{
			name:         "negative number",
			x:            NewInt(-100),
			wantResult:   "100",
			wantOverflow: false,
		},
		{
			name:         "max_int256",
			x:            MustFromDecimal("57896044618658097711785492504343953926634992332820282019728792003956564819967"),
			wantResult:   "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			wantOverflow: false,
		},
		{
			name:         "min_int256",
			x:            MustFromDecimal("-57896044618658097711785492504343953926634992332820282019728792003956564819968"),
			wantResult:   "-57896044618658097711785492504343953926634992332820282019728792003956564819968",
			wantOverflow: true,
		},
		{
			name:         "min_int256 + 1",
			x:            MustFromDecimal("-57896044618658097711785492504343953926634992332820282019728792003956564819967"),
			wantResult:   "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			wantOverflow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := new(Int)
			gotResult, gotOverflow := z.AbsOverflow(tt.x)

			if gotOverflow != tt.wantOverflow {
				t.Errorf("overflow = %v, want %v", gotOverflow, tt.wantOverflow)
				return
			}

			if gotResult == nil {
				t.Error("unexpected nil result")
				return
			}

			if gotResult.ToString() != tt.wantResult {
				t.Errorf("result = %v, want %v", gotResult.ToString(), tt.wantResult)
			}

			// abs value must be GTE 0 (if there is no overflow)
			if !tt.wantOverflow && gotResult.neg && !gotResult.IsZero() {
				t.Error("absolute value cannot be negative")
			}

			// original value must not be modified
			originalValue := tt.x.ToString()
			if tt.x.ToString() != originalValue {
				t.Errorf("original value was modified: got %v, want %v",
					tt.x.ToString(), originalValue)
			}
		})
	}
}
