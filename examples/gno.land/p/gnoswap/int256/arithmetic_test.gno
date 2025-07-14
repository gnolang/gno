package int256

import (
	"testing"

	"gno.land/p/gnoswap/uint256"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "2", "3"},
		// NEGATIVE
		{"-1", "1", "0"},
		{"1", "-1", "0"},
		{"3", "-3", "0"},
		{"-1", "-1", "-2"},
		{"-1", "-2", "-3"},
		{"-1", "3", "2"},
		{"3", "-1", "2"},
		// OVERFLOW
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "1", "0"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Add(x, y)

		if got.Neq(want) {
			t.Errorf("Add(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestAddUint256(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "2", "3"},
		{"-1", "1", "0"},
		{"-1", "3", "2"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639934", "115792089237316195423570985008687907853269984665640564039457584007913129639935", "1"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639934", "-1"},
		// OVERFLOW
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "115792089237316195423570985008687907853269984665640564039457584007913129639935", "0"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.AddUint256(x, y)

		if got.Neq(want) {
			t.Errorf("AddUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestAddDelta(t *testing.T) {
	tests := []struct {
		z, x, y, want string
	}{
		{"0", "0", "0", "0"},
		{"0", "0", "1", "1"},
		{"0", "1", "0", "1"},
		{"0", "1", "1", "2"},
		{"1", "2", "3", "5"},
		{"5", "10", "-3", "7"},
		// underflow
		{"1", "2", "-3", "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range tests {
		z, err := uint256.FromDecimal(tc.z)
		if err != nil {
			t.Error(err)
			continue
		}

		x, err := uint256.FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := uint256.FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		AddDelta(z, x, y)

		if z.Neq(want) {
			t.Errorf("AddDelta(%s, %s, %s) = %v, want %v", tc.z, tc.x, tc.y, z.ToString(), want.ToString())
		}
	}
}

func TestAddDeltaOverflow(t *testing.T) {
	tests := []struct {
		z, x, y string
		want    bool
	}{
		{"0", "0", "0", false},
		// underflow
		{"1", "2", "-3", true},
	}

	for _, tc := range tests {
		z, err := uint256.FromDecimal(tc.z)
		if err != nil {
			t.Error(err)
			continue
		}

		x, err := uint256.FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := FromDecimal(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		result := AddDeltaOverflow(z, x, y)
		if result != tc.want {
			t.Errorf("AddDeltaOverflow(%s, %s, %s) = %v, want %v", tc.z, tc.x, tc.y, result, tc.want)
		}
	}
}

func TestAddOverflow(t *testing.T) {
	maxInt256 := MaxInt256()
	minInt256 := MinInt256()

	tests := []struct {
		name         string
		x            *Int
		y            *Int
		wantResult   string
		wantOverflow bool
	}{
		// Basic cases (no overflow)
		{
			name:         "positive + positive (no overflow)",
			x:            NewInt(100),
			y:            NewInt(200),
			wantResult:   "300",
			wantOverflow: false,
		},
		{
			name:         "negative + negative (no overflow)",
			x:            NewInt(-100),
			y:            NewInt(-200),
			wantResult:   "-300",
			wantOverflow: false,
		},
		// Boundary cases - near maximum value
		{
			name:         "max_int256 + 0",
			x:            maxInt256,
			y:            Zero(),
			wantResult:   maxInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "max_int256 - 1 + 1",
			x:            new(Int).Sub(maxInt256, One()),
			y:            One(),
			wantResult:   maxInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "max_int256 + 1",
			x:            maxInt256,
			y:            One(),
			wantResult:   "", // overflow
			wantOverflow: true,
		},

		// Boundary cases - near minimum value
		{
			name:         "min_int256 + 0",
			x:            minInt256,
			y:            Zero(),
			wantResult:   minInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "min_int256 + 1 - 1",
			x:            new(Int).Add(minInt256, One()),
			y:            NewInt(-1),
			wantResult:   minInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "min_int256 + (-1)",
			x:            minInt256,
			y:            NewInt(-1),
			wantResult:   "", // overflow
			wantOverflow: true,
		},

		// Special cases
		{
			name:         "max_int256 + min_int256",
			x:            maxInt256,
			y:            minInt256,
			wantResult:   "-1",
			wantOverflow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := new(Int)
			gotResult, gotOverflow := z.AddOverflow(tt.x, tt.y)

			if gotOverflow != tt.wantOverflow {
				t.Errorf("overflow = %v, want %v", gotOverflow, tt.wantOverflow)
				return
			}

			if !gotOverflow {
				if gotResult == nil {
					t.Error("unexpected nil result for non-overflow case")
					return
				}
				if gotResult.ToString() != tt.wantResult {
					t.Errorf("result = %v, want %v", gotResult.ToString(), tt.wantResult)
				}
			}

			// Commutativity test only for non-overflow cases
			if !tt.wantOverflow {
				reverseResult, reverseOverflow := z.AddOverflow(tt.y, tt.x)
				if reverseOverflow != gotOverflow {
					t.Error("addition is not commutative for overflow")
				}
				if reverseResult.ToString() != gotResult.ToString() {
					t.Error("addition is not commutative for result")
				}
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"-1", "1", "-2"},
		{"1", "-1", "2"},
		{"-1", "-1", "0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0", "-115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{x: "-115792089237316195423570985008687907853269984665640564039457584007913129639935", y: "1", want: "0"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Sub(x, y)

		if got.Neq(want) {
			t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestSubUint256(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "-1"},
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"1", "2", "-1"},
		{"-1", "1", "-2"},
		{"-1", "3", "-4"},
		// underflow
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "1", "-0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "2", "-1"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "3", "-2"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.SubUint256(x, y)

		if got.Neq(want) {
			t.Errorf("SubUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"5", "3", "15"},
		{"-5", "3", "-15"},
		{"5", "-3", "-15"},
		{"0", "3", "0"},
		{"3", "0", "0"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Mul(x, y)

		if got.Neq(want) {
			t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestInt_SubOverflow(t *testing.T) {
	maxInt256 := MaxInt256()
	minInt256 := MinInt256()

	tests := []struct {
		name         string
		x            *Int
		y            *Int
		wantResult   string
		wantOverflow bool
	}{
		{
			name:         "positive - positive (no overflow)",
			x:            NewInt(200),
			y:            NewInt(100),
			wantResult:   "100",
			wantOverflow: false,
		},
		{
			name:         "negative - negative (no overflow)",
			x:            NewInt(-200),
			y:            NewInt(-300),
			wantResult:   "100",
			wantOverflow: false,
		},
		{
			name:         "positive - negative (no overflow)",
			x:            NewInt(200),
			y:            NewInt(-100),
			wantResult:   "300",
			wantOverflow: false,
		},
		{
			name:         "max_int256 - 0",
			x:            maxInt256,
			y:            Zero(),
			wantResult:   maxInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "min_int256 - 0",
			x:            minInt256,
			y:            Zero(),
			wantResult:   minInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "max_int256 - (-1)", // max_int256 + 1 -> overflow
			x:            maxInt256,
			y:            NewInt(-1),
			wantResult:   "",
			wantOverflow: true,
		},
		{
			name:         "min_int256 - 1", // min_int256 - 1 -> overflow
			x:            minInt256,
			y:            One(),
			wantResult:   "",
			wantOverflow: true,
		},
		{
			name:         "0 - 0",
			x:            Zero(),
			y:            Zero(),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "min_int256 - min_int256",
			x:            minInt256,
			y:            minInt256,
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "max_int256 - max_int256",
			x:            maxInt256,
			y:            maxInt256,
			wantResult:   "0",
			wantOverflow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// z := new(Int)
			z := New()
			gotResult, gotOverflow := z.SubOverflow(tt.x, tt.y)

			if gotOverflow != tt.wantOverflow {
				t.Errorf("overflow = %v, want %v", gotOverflow, tt.wantOverflow)
				return
			}

			if !gotOverflow {
				if gotResult == nil {
					t.Error("unexpected nil result for non-overflow case")
					return
				}
				if gotResult.ToString() != tt.wantResult {
					t.Errorf("result = %v, want %v", gotResult.ToString(), tt.wantResult)
				}
			}
		})
	}
}

func TestMulUint256(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"1", "2", "2"},
		{"-1", "1", "-1"},
		{"-1", "3", "-3"},
		{"3", "4", "12"},
		{"-3", "4", "-12"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639934", "2", "-115792089237316195423570985008687907853269984665640564039457584007913129639932"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639934", "2", "115792089237316195423570985008687907853269984665640564039457584007913129639932"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.MulUint256(x, y)

		if got.Neq(want) {
			t.Errorf("MulUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestInt_MulOverflow(t *testing.T) {
	maxInt256 := MaxInt256()
	minInt256 := MinInt256()

	tests := []struct {
		name         string
		x            *Int
		y            *Int
		wantResult   string
		wantOverflow bool
	}{
		{
			name:         "positive * positive (no overflow)",
			x:            NewInt(100),
			y:            NewInt(100),
			wantResult:   "10000",
			wantOverflow: false,
		},
		{
			name:         "negative * negative (no overflow)",
			x:            NewInt(-100),
			y:            NewInt(-100),
			wantResult:   "10000",
			wantOverflow: false,
		},
		{
			name:         "positive * negative (no overflow)",
			x:            NewInt(100),
			y:            NewInt(-100),
			wantResult:   "-10000",
			wantOverflow: false,
		},
		{
			name:         "0 * positive",
			x:            Zero(),
			y:            NewInt(100),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "positive * 0",
			x:            NewInt(100),
			y:            Zero(),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "0 * 0",
			x:            Zero(),
			y:            Zero(),
			wantResult:   "0",
			wantOverflow: false,
		},
		{
			name:         "max_int256 * 1",
			x:            maxInt256,
			y:            One(),
			wantResult:   maxInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "min_int256 * 1",
			x:            minInt256,
			y:            One(),
			wantResult:   minInt256.ToString(),
			wantOverflow: false,
		},
		{
			name:         "min_int256 * -1",
			x:            minInt256,
			y:            NewInt(-1),
			wantResult:   "", // overflow because abs(min_int256) > max_int256
			wantOverflow: true,
		},
		{
			name:         "max_int256 * 2",
			x:            maxInt256,
			y:            NewInt(2),
			wantResult:   "",
			wantOverflow: true,
		},
		{
			name:         "min_int256 * 2",
			x:            minInt256,
			y:            NewInt(2),
			wantResult:   "",
			wantOverflow: true,
		},
		{
			name:         "half_max * 2",
			x:            MustFromDecimal("28948022309329048855892746252171976963317496332820282019728792003956564819983"), // (2^255-1)/2
			y:            NewInt(2),
			wantResult:   "",
			wantOverflow: true,
		},
		{
			name:         "(half_max + 1) * 2",
			x:            new(Int).Add(new(Int).Div(maxInt256, NewInt(2)), One()),
			y:            NewInt(2),
			wantResult:   "",
			wantOverflow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := new(Int)
			gotResult, gotOverflow := z.MulOverflow(tt.x, tt.y)

			if gotOverflow != tt.wantOverflow {
				t.Errorf("overflow = %v, want %v", gotOverflow, tt.wantOverflow)
				return
			}

			if !gotOverflow {
				if gotResult == nil {
					t.Error("unexpected nil result for non-overflow case")
					return
				}
				if gotResult.ToString() != tt.wantResult {
					t.Errorf("result = %v, want %v", gotResult.ToString(), tt.wantResult)
				}
			}

			if !tt.wantOverflow {
				reverseResult, reverseOverflow := z.MulOverflow(tt.y, tt.x)
				if reverseOverflow != gotOverflow {
					t.Error("multiplication is not commutative for overflow")
				}
				if reverseResult.ToString() != gotResult.ToString() {
					t.Error("multiplication is not commutative for result")
				}
			}
		})
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		x, y, expected string
	}{
		{"1", "1", "1"},
		{"0", "1", "0"},
		{"-1", "1", "-1"},
		{"1", "-1", "-1"},
		{"-1", "-1", "1"},
		{"-6", "3", "-2"},
		{"10", "-2", "-5"},
		{"-10", "3", "-3"},
		{"7", "3", "2"},
		{"-7", "3", "-2"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", "2", "57896044618658097711785492504343953926634992332820282019728792003956564819967"}, // Max uint256 / 2
	}

	for _, tt := range tests {
		t.Run(tt.x+"/"+tt.y, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			y := MustFromDecimal(tt.y)
			result := Zero().Div(x, y)
			if result.ToString() != tt.expected {
				t.Errorf("Div(%s, %s) = %s, want %s", tt.x, tt.y, result.ToString(), tt.expected)
			}
			if result.abs.IsZero() && result.neg {
				t.Errorf("Div(%s, %s) resulted in negative zero", tt.x, tt.y)
			}
		})
	}

	t.Run("Division by zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Div(1, 0) did not panic")
			}
		}()
		x := MustFromDecimal("1")
		y := MustFromDecimal("0")
		Zero().Div(x, y)
	})
}

func TestDivUint256(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"1", "2", "0"},
		{"-1", "1", "-1"},
		{"-1", "3", "0"},
		{"4", "3", "1"},
		{"25", "5", "5"},
		{"25", "4", "6"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639934", "2", "-57896044618658097711785492504343953926634992332820282019728792003956564819967"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639934", "2", "57896044618658097711785492504343953926634992332820282019728792003956564819967"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.DivUint256(x, y)

		if got.Neq(want) {
			t.Errorf("DivUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestQuo(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"0", "-1", "0"},
		{"10", "1", "10"},
		{"10", "-1", "-10"},
		{"-10", "1", "-10"},
		{"-10", "-1", "10"},
		{"10", "-3", "-3"},
		{"10", "3", "3"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Quo(x, y)

		if got.Neq(want) {
			t.Errorf("Quo(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestRem(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"0", "-1", "0"},
		{"10", "1", "0"},
		{"10", "-1", "0"},
		{"-10", "1", "0"},
		{"-10", "-1", "0"},
		{"10", "3", "1"},
		{"10", "-3", "1"},
		{"-10", "3", "-1"},
		{"-10", "-3", "-1"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Rem(x, y)

		if got.Neq(want) {
			t.Errorf("Rem(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestMod(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"0", "-1", "0"},
		{"10", "0", "0"},
		{"10", "1", "0"},
		{"10", "-1", "0"},
		{"-10", "0", "0"},
		{"-10", "1", "0"},
		{"-10", "-1", "0"},
		{"10", "3", "1"},
		{"10", "-3", "1"},
		{"-10", "3", "2"},
		{"-10", "-3", "2"},
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

		want, err := FromDecimal(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Mod(x, y)

		if got.Neq(want) {
			t.Errorf("Mod(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
}

func TestNilInitialization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*Int, error)
		wantStr string
	}{
		{
			name: "AddUint256 with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := uint256.NewUint(5)
				return z.AddUint256(z, x), nil
			},
			wantStr: "5",
		},
		{
			name: "SubUint256 with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := uint256.NewUint(5)
				return z.SubUint256(z, x), nil
			},
			wantStr: "-5",
		},
		{
			name: "MulUint256 with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := uint256.NewUint(5)
				return z.MulUint256(z, x), nil
			},
			wantStr: "0",
		},
		{
			name: "DivUint256 with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := uint256.NewUint(5)
				return z.DivUint256(z, x), nil
			},
			wantStr: "0",
		},
		{
			name: "Mod with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := MustFromDecimal("5")
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Mod with nil abs panicked: %v", r)
					}
				}()
				return z.Mod(z, x), nil
			},
			wantStr: "0",
		},
		{
			name: "Chained operations with nil abs",
			setup: func() (*Int, error) {
				z := new(Int)
				x := uint256.NewUint(5)
				y := uint256.NewUint(3)
				// (0 + 5) * 3
				return z.AddUint256(z, x).MulUint256(z, y), nil
			},
			wantStr: "15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := result.ToString()
			if got != tt.wantStr {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.wantStr)
			}
		})
	}
}
