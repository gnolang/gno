package int256

import (
	"testing"

	"gno.land/p/demo/uint256"
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
		{"-1", "1", "-0"}, // TODO: remove negative sign for 0 ??
		{"1", "-1", "0"},
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

func TestSub(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"-1", "1", "-2"},
		{"1", "-1", "2"},
		{"-1", "-1", "-0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-115792089237316195423570985008687907853269984665640564039457584007913129639935", "-0"},
		{"-115792089237316195423570985008687907853269984665640564039457584007913129639935", "0", "-115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{x: "-115792089237316195423570985008687907853269984665640564039457584007913129639935", y: "1", want: "-0"},
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

func TestDiv(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1", "0"},
		{"0", "-1", "-0"},
		{"10", "0", "0"},
		{"10", "1", "10"},
		{"10", "-1", "-10"},
		{"-10", "0", "-0"},
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
		got.Div(x, y)

		if got.Neq(want) {
			t.Errorf("Div(%s, %s) = %v, want %v", tc.x, tc.y, got.ToString(), want.ToString())
		}
	}
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
