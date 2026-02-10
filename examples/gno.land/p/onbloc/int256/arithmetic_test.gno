package int256

import (
	"testing"

	"gno.land/p/onbloc/uint256"
)

const (
	// 2^255 - 1
	MAX_INT256 = "57896044618658097711785492504343953926634992332820282019728792003956564819967"
	// -(2^255 - 1)
	MINUS_MAX_INT256 = "-57896044618658097711785492504343953926634992332820282019728792003956564819967"

	// 2^255 - 1
	MAX_UINT256         = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
	MAX_UINT256_MINUS_1 = "115792089237316195423570985008687907853269984665640564039457584007913129639934"

	MINUS_MAX_UINT256        = "-115792089237316195423570985008687907853269984665640564039457584007913129639935"
	MINUS_MAX_UINT256_PLUS_1 = "-115792089237316195423570985008687907853269984665640564039457584007913129639934"

	TWO_POW_128               = "340282366920938463463374607431768211456"
	MINUS_TWO_POW_128         = "-340282366920938463463374607431768211456"
	MINUS_TWO_POW_128_MINUS_1 = "-340282366920938463463374607431768211457"
	TWO_POW_128_MINUS_1       = "340282366920938463463374607431768211455"

	TWO_POW_129_MINUS_1 = "680564733841876926926749214863536422911"

	TWO_POW_254           = "28948022309329048855892746252171976963317496166410141009864396001978282409984"
	MINUS_TWO_POW_254     = "-28948022309329048855892746252171976963317496166410141009864396001978282409984"
	HALF_MAX_INT256       = "28948022309329048855892746252171976963317496166410141009864396001978282409983"
	MINUS_HALF_MAX_INT256 = "-28948022309329048855892746252171976963317496166410141009864396001978282409983"

	TWO_POW_255        = "57896044618658097711785492504343953926634992332820282019728792003956564819968"
	MIN_INT256         = "-57896044618658097711785492504343953926634992332820282019728792003956564819968"
	MIN_INT256_MINUS_1 = "-57896044618658097711785492504343953926634992332820282019728792003956564819969"
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
		{MAX_UINT256, "1", "0"},
		{MAX_INT256, "1", MIN_INT256},
		{MIN_INT256, "-1", MAX_INT256},
		{MAX_INT256, MAX_INT256, "-2"},
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
			t.Errorf("Add(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
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
		{MINUS_MAX_UINT256_PLUS_1, MAX_UINT256, "1"},
		{MINUS_MAX_UINT256, MAX_UINT256_MINUS_1, "-1"},
		// OVERFLOW
		{MINUS_MAX_UINT256, MAX_UINT256, "0"},
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
			t.Errorf("AddUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
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
		{"1", "2", "-3", MAX_UINT256},
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
			t.Errorf("AddDelta(%s, %s, %s) = %v, want %v", tc.z, tc.x, tc.y, z.String(), want.String())
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

func TestSub(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"-1", "1", "-2"},
		{"1", "-1", "2"},
		{"-1", "-1", "0"},
		{MINUS_MAX_UINT256, MINUS_MAX_UINT256, "0"},
		{MINUS_MAX_UINT256, "0", MINUS_MAX_UINT256},
		{MAX_INT256, MIN_INT256, "-1"},
		{MIN_INT256, MIN_INT256, "0"},
		{MAX_INT256, MAX_INT256, "0"},
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
			t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
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
		{MINUS_MAX_UINT256, "1", "0"},
		{MINUS_MAX_UINT256, "2", "-1"},
		{MINUS_MAX_UINT256, "3", "-2"},
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
			t.Errorf("SubUint256(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
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
		{"-5", "-3", "15"},
		{MAX_UINT256, "1", MAX_UINT256},
		{MAX_INT256, "2", "-2"},
		{TWO_POW_254, "2", MIN_INT256},
		{MINUS_TWO_POW_254, "2", MIN_INT256},
		{MAX_INT256, "1", MAX_INT256},
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
			t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
		}
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
		// the maximum value of a positive number in int256 is less than the maximum value of a uint256
		{MAX_INT256, "2", HALF_MAX_INT256},
		{MINUS_MAX_INT256, "2", MINUS_HALF_MAX_INT256},
		{MAX_INT256, "-1", MINUS_MAX_INT256},
	}

	for _, tt := range tests {
		t.Run(tt.x+"/"+tt.y, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			y := MustFromDecimal(tt.y)
			result := Zero().Div(x, y)
			if result.String() != tt.expected {
				t.Errorf("Div(%s, %s) = %s, want %s", tt.x, tt.y, result.String(), tt.expected)
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
		{"-10", "3", "-3"},
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
			t.Errorf("Quo(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
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
			t.Errorf("Rem(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
		}
	}
}

func TestMod(t *testing.T) {
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
			t.Errorf("Mod(%s, %s) = %v, want %v", tc.x, tc.y, got.String(), want.String())
		}
	}
}

func TestModeOverflow(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{MIN_INT256, "2", "0"},  // MIN_INT256 % 2 = 0
		{MAX_INT256, "2", "1"},  // MAX_INT256 % 2 = 1
		{MIN_INT256, "-1", "0"}, // MIN_INT256 % -1 = 0
		{MAX_INT256, "-1", "0"}, // MAX_INT256 % -1 = 0
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)
		got := New().Mod(x, y)
		if got.Neq(want) {
			t.Errorf("Mod(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestModPanic(t *testing.T) {
	tests := []struct {
		x, y string
	}{
		{"10", "0"},
		{"10", "-0"},
		{"-10", "0"},
		{"-10", "-0"},
	}

	for _, tc := range tests {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Mod(%s, %s) did not panic", tc.x, tc.y)
			}
		}()
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

		result := New().Mod(x, y)
		t.Errorf("Mod(%s, %s) = %v, want %v", tc.x, tc.y, result.String(), "0")
	}
}

func TestDivE(t *testing.T) {
	testCases := []struct {
		x, y int64
		want int64
	}{
		{8, 3, 2},
		{8, -3, -2},
		{-8, 3, -3},
		{-8, -3, 3},
		{1, 2, 0},
		{1, -2, 0},
		{-1, 2, -1},
		{-1, -2, 1},
		{0, 1, 0},
		{0, -1, 0},
	}

	for _, tc := range testCases {
		x := NewInt(tc.x)
		y := NewInt(tc.y)
		want := NewInt(tc.want)
		got := new(Int).DivE(x, y)
		if got.Cmp(want) != 0 {
			t.Errorf("DivE(%v, %v) = %v, want %v", tc.x, tc.y, got, want)
		}
	}
}

func TestDivEByZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("DivE did not panic on division by zero")
		}
	}()

	x := NewInt(1)
	y := NewInt(0)
	new(Int).DivE(x, y)
}

func TestModEByZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ModE did not panic on division by zero")
		}
	}()

	x := NewInt(1)
	y := NewInt(0)
	new(Int).ModE(x, y)
}

func TestLargeNumbers(t *testing.T) {
	x, _ := new(Int).SetString("123456789012345678901234567890")
	y, _ := new(Int).SetString("987654321098765432109876543210")

	// Expected results (calculated separately)
	expectedQ, _ := new(Int).SetString("0")
	expectedR, _ := new(Int).SetString("123456789012345678901234567890")

	gotQ := new(Int).DivE(x, y)
	gotR := new(Int).ModE(x, y)

	if gotQ.Cmp(expectedQ) != 0 {
		t.Errorf("DivE with large numbers: got %v, want %v", gotQ, expectedQ)
	}

	if gotR.Cmp(expectedR) != 0 {
		t.Errorf("ModE with large numbers: got %v, want %v", gotR, expectedR)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		x, want string
	}{
		{"0", "0"},
		{"1", "1"},
		{"-1", "1"},
		{"-2", "2"},
		{"-100000000000", "100000000000"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.Abs()

		if got.String() != tc.want {
			t.Errorf("Abs(%s) = %v, want %v", tc.x, got.String(), tc.want)
		}
	}
}
