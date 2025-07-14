package int256

import (
	"testing"

	"gno.land/p/gnoswap/uint256"
)

func TestSetInt64(t *testing.T) {
	tests := []struct {
		x    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{9223372036854775807, "9223372036854775807"},
		{-9223372036854775808, "-9223372036854775808"},
	}

	for _, tc := range tests {
		var z Int
		z.SetInt64(tc.x)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("SetInt64(%d) = %s, want %s", tc.x, got, tc.want)
		}
	}
}

func TestSetUint64(t *testing.T) {
	tests := []struct {
		x    uint64
		want string
	}{
		{0, "0"},
		{1, "1"},
	}

	for _, tc := range tests {
		var z Int
		z.SetUint64(tc.x)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("SetUint64(%d) = %s, want %s", tc.x, got, tc.want)
		}
	}
}

func TestUint64(t *testing.T) {
	tests := []struct {
		x    string
		want uint64
	}{
		{"0", 0},
		{"1", 1},
		{"9223372036854775807", 9223372036854775807},
		{"9223372036854775808", 9223372036854775808},
		{"18446744073709551615", 18446744073709551615},
		{"18446744073709551616", 0},
		{"18446744073709551617", 1},
		{"-1", 1},
		{"-18446744073709551615", 18446744073709551615},
		{"-18446744073709551616", 0},
		{"-18446744073709551617", 1},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)

		got := z.Uint64()
		if got != tc.want {
			t.Errorf("Uint64(%s) = %d, want %d", tc.x, got, tc.want)
		}
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		x    string
		want int64
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
		{"9223372036854775807", 9223372036854775807},
		{"-9223372036854775808", -9223372036854775808},
		{"9223372036854775808", -9223372036854775808},
		{"-9223372036854775809", 9223372036854775807},
		{"18446744073709551616", 0},
		{"18446744073709551617", 1},
		{"18446744073709551615", -1},
		{"-18446744073709551615", 1},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)

		got := z.Int64()
		if got != tc.want {
			t.Errorf("Int64(%s) = %d, want %d", tc.x, got, tc.want)
		}
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "-1"},
		{"-1", "1"},
		{"9223372036854775807", "-9223372036854775807"},
		{"-18446744073709551615", "18446744073709551615"},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)
		z.Neg(z)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("Neg(%s) = %s, want %s", tc.x, got, tc.want)
		}
	}
}

func TestInt_NegOverflow(t *testing.T) {
	maxInt256 := MaxInt256()
	minInt256 := MinInt256()

	negMaxInt256 := New().Neg(maxInt256)

	tests := []struct {
		name         string
		x            *Int
		wantResult   string
		wantOverflow bool
	}{
		{
			name:         "negate zero",
			x:            Zero(),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "negate positive",
			x:            NewInt(100),
			wantResult:   "-100",
			wantOverflow: false,
		},
		{
			name:         "negate negative",
			x:            NewInt(-100),
			wantResult:   "100",
			wantOverflow: false,
		},
		{
			name:         "negate max_int256",
			x:            maxInt256,
			wantResult:   negMaxInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "negate min_int256",
			x:            minInt256,
			wantResult:   minInt256.ToString(), // must preserve the original value
			wantOverflow: true,
		},
		{
			name:         "negate (min_int256 + 1)",
			x:            new(Int).Add(minInt256, One()),
			wantResult:   new(Int).Sub(maxInt256, Zero()).ToString(),
			wantOverflow: false,
		},
		{
			name:         "negate (max_int256 - 1)",
			x:            MustFromDecimal("57896044618658097711785492504343953926634992332820282019728792003956564819966"), // max_int256 - 1
			wantResult:   "-57896044618658097711785492504343953926634992332820282019728792003956564819966",
			wantOverflow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := new(Int)
			gotResult, gotOverflow := z.NegOverflow(tt.x)

			if gotOverflow != tt.wantOverflow {
				t.Errorf("overflow = %v, want %v", gotOverflow, tt.wantOverflow)
				return
			}

			if gotResult == nil {
				t.Error("unexpected nil result")
				return
			}

			if gotResult.ToString() != tt.wantResult {
				// use almost equal comparison to handle the precision issue
				diff := new(Int).Sub(gotResult, MustFromDecimal(tt.wantResult))
				if diff.Abs().Cmp(uint256.NewUint(1)) > 0 {
					t.Errorf("result = %v, want %v", gotResult.ToString(), tt.wantResult)
				}
			}

			// double negation test (only if there is no overflow)
			if !tt.wantOverflow {
				doubleNegResult, doubleNegOverflow := new(Int).NegOverflow(gotResult)
				if doubleNegOverflow {
					t.Error("unexpected overflow in double negation")
				}
				if doubleNegResult.ToString() != tt.x.ToString() {
					t.Errorf("double negation result = %v, want %v",
						doubleNegResult.ToString(), tt.x.ToString())
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"9223372036854775807", "9223372036854775807"},
		{"-18446744073709551615", "-18446744073709551615"},
	}

	for _, tc := range tests {
		z := MustFromDecimal(tc.x)
		z.Set(z)

		got := z.ToString()
		if got != tc.want {
			t.Errorf("set(%s) = %s, want %s", tc.x, got, tc.want)
		}
	}
}

func TestSetUint256(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"9223372036854775807", "9223372036854775807"},
		{"18446744073709551615", "18446744073709551615"},
	}

	for _, tc := range tests {
		got := New()

		z := uint256.MustFromDecimal(tc.x)
		got.SetUint256(z)

		if got.ToString() != tc.want {
			t.Errorf("SetUint256(%s) = %s, want %s", tc.x, got.ToString(), tc.want)
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Int
		expected string
	}{
		{
			name: "Zero from subtraction",
			setup: func() *Int {
				minusThree := MustFromDecimal("-3")
				three := MustFromDecimal("3")
				return Zero().Add(minusThree, three)
			},
			expected: "0",
		},
		{
			name: "Zero from right shift",
			setup: func() *Int {
				return Zero().Rsh(One(), 1234)
			},
			expected: "0",
		},
		{
			name: "Positive number",
			setup: func() *Int {
				return MustFromDecimal("42")
			},
			expected: "42",
		},
		{
			name: "Negative number",
			setup: func() *Int {
				return MustFromDecimal("-42")
			},
			expected: "-42",
		},
		{
			name: "Large positive number",
			setup: func() *Int {
				return MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935")
			},
			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		{
			name: "Large negative number",
			setup: func() *Int {
				return MustFromDecimal("-115792089237316195423570985008687907853269984665640564039457584007913129639935")
			},
			expected: "-115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := tt.setup()
			result := z.ToString()
			if result != tt.expected {
				t.Errorf("ToString() = %s, want %s", result, tt.expected)
			}
		})
	}
}
