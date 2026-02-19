package uint256

import (
	"testing"
)

type binOp2Test struct {
	x, y, want string
}

func TestAdd(t *testing.T) {
	tests := []binOp2Test{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "3", "4"},
		{"10", "10", "20"},
		{"18446744073709551615", "18446744073709551615", "36893488147419103230"}, // uint64 overflow
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)

		want := MustFromDecimal(tt.want)
		got := new(Uint).Add(x, y)

		if got.Neq(want) {
			t.Errorf("Add(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestAddOverflow(t *testing.T) {
	tests := []struct {
		x, y     string
		want     string
		overflow bool
	}{
		{"0", "1", "1", false},
		{"1", "0", "1", false},
		{"1", "1", "2", false},
		{"10", "10", "20", false},
		{"18446744073709551615", "18446744073709551615", "36893488147419103230", false}, // uint64 overflow, but not Uint256 overflow
		{twoPow256Sub1, "1", "0", true}, // 2^256 - 1 + 1, should overflow
		{"57896044618658097711785492504343953926634992332820282019728792003956564819967", "57896044618658097711785492504343953926634992332820282019728792003956564819968", twoPow256Sub1, false}, // (2^255 - 1) + 2^255, no overflow
		{"57896044618658097711785492504343953926634992332820282019728792003956564819967", "57896044618658097711785492504343953926634992332820282019728792003956564819969", "0", true},            // (2^255 - 1) + (2^255 + 1), should overflow
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want, _ := FromDecimal(tt.want)

		got, overflow := new(Uint).AddOverflow(x, y)

		if got.Cmp(want) != 0 || overflow != tt.overflow {
			t.Errorf("AddOverflow(%s, %s) = (%s, %v), want (%s, %v)",
				tt.x, tt.y, got.String(), overflow, tt.want, tt.overflow)
		}
	}
}

func TestSub(t *testing.T) {
	tests := []binOp2Test{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"10", "10", "0"},
		{"31337", "1337", "30000"},
		{"2", "3", twoPow256Sub1}, // underflow
	}

	for _, tc := range tests {
		x := MustFromDecimal(tc.x)
		y := MustFromDecimal(tc.y)

		want := MustFromDecimal(tc.want)

		got := new(Uint).Sub(x, y)

		if got.Neq(want) {
			t.Errorf(
				"Sub(%s, %s) = %v, want %v",
				tc.x, tc.y, got.String(), want.String(),
			)
		}
	}
}

func TestSubOverflow(t *testing.T) {
	tests := []struct {
		x, y     string
		want     string
		overflow bool
	}{
		{"1", "0", "1", false},
		{"1", "1", "0", false},
		{"10", "10", "0", false},
		{"31337", "1337", "30000", false},
		{"0", "1", twoPow256Sub1, true}, // 0 - 1, should underflow
		{"57896044618658097711785492504343953926634992332820282019728792003956564819968", "1", "57896044618658097711785492504343953926634992332820282019728792003956564819967", false},          // 2^255 - 1, no underflow
		{"57896044618658097711785492504343953926634992332820282019728792003956564819968", "57896044618658097711785492504343953926634992332820282019728792003956564819969", twoPow256Sub1, true}, // 2^255 - (2^255 + 1), should underflow
	}

	for _, tc := range tests {
		x := MustFromDecimal(tc.x)
		y := MustFromDecimal(tc.y)
		want := MustFromDecimal(tc.want)

		got, overflow := new(Uint).SubOverflow(x, y)

		if got.Cmp(want) != 0 || overflow != tc.overflow {
			t.Errorf(
				"SubOverflow(%s, %s) = (%s, %v), want (%s, %v)",
				tc.x, tc.y, got.String(), overflow, tc.want, tc.overflow,
			)
		}
	}
}

func TestMul(t *testing.T) {
	tests := []binOp2Test{
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"10", "10", "100"},
		{"18446744073709551615", "2", "36893488147419103230"}, // uint64 overflow
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)
		got := new(Uint).Mul(x, y)

		if got.Neq(want) {
			t.Errorf("Mul(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestMulOverflow(t *testing.T) {
	tests := []struct {
		x        string
		y        string
		wantZ    string
		wantOver bool
	}{
		{"0x1", "0x1", "0x1", false},
		{"0x0", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0x0", false},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0x2", "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe", true},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0x1", true},
		{"0x8000000000000000000000000000000000000000000000000000000000000000", "0x2", "0x0", true},
		{"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0x2", "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe", false},
		{"0x100000000000000000", "0x100000000000000000", "0x10000000000000000000000000000000000", false},
		{"0x10000000000000000000000000000000", "0x10000000000000000000000000000000", "0x100000000000000000000000000000000000000000000000000000000000000", false},
	}

	for _, tt := range tests {
		x := MustFromHex(tt.x)
		y := MustFromHex(tt.y)
		wantZ := MustFromHex(tt.wantZ)

		gotZ, gotOver := new(Uint).MulOverflow(x, y)

		if gotZ.Neq(wantZ) {
			t.Errorf(
				"MulOverflow(%s, %s) = %s, want %s",
				tt.x, tt.y, gotZ.String(), wantZ.String(),
			)
		}
		if gotOver != tt.wantOver {
			t.Errorf("MulOverflow(%s, %s) = %v, want %v", tt.x, tt.y, gotOver, tt.wantOver)
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "10445"},
		{"31337", "0", "0"},
		{"0", "31337", "0"},
		{"1", "1", "1"},
		{"1000000000000000000", "3", "333333333333333333"},
		{twoPow256Sub1, "2", "57896044618658097711785492504343953926634992332820282019728792003956564819967"},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).Div(x, y)

		if got.Neq(want) {
			t.Errorf("Div(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestSDiv(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "10445"},
		{"1000000000000000000", "3", "333333333333333333"},
		{"1", "1", "1"},
		{"0", "31337", "0"},
		{"31337", "0", "0"},
		{twoPow256Sub1, "1", twoPow256Sub1}, // -1 / 1 = -1
		{twoPow256Sub1, twoPow256Sub1, "1"}, // -1 / -1 = 1
		{"2", twoPow256Sub1, "115792089237316195423570985008687907853269984665640564039457584007913129639934"}, // 2 / -1 = -2
		{twoPow256Sub1, "2", "0"},                           // -1 / 2 = 0
		{twoPow256Sub1, twoPow256Sub1, "1"},                 // max uint256 / max uint256
		{"0", twoPow256Sub1, "0"},                           // 0 / -1
		{"9223372036854775808", "2", "4611686018427387904"}, // big num / 2
		{"9223372036854775808", twoPow256Sub1, "115792089237316195423570985008687907853269984665640564039448360635876274864128"}, // big / -1 = - big num
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).SDiv(x, y)

		if got.Neq(want) {
			t.Errorf("SDiv(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestMod(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "2"},
		{"31337", "0", "0"},
		{"0", "31337", "0"},
		{"2", "31337", "2"},
		{"1", "1", "0"},
		{twoPow256Sub1, "2", "1"}, // 2^256 - 1 mod 2
		{twoPow256Sub1, "3", "0"}, // 2^256 - 1 mod 3
		{twoPow256Sub1, "57896044618658097711785492504343953926634992332820282019728792003956564819968", "57896044618658097711785492504343953926634992332820282019728792003956564819967"}, // 2^256 - 1 mod 2^255
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).Mod(x, y)

		if got.Neq(want) {
			t.Errorf("Mod(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestSMod(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "2"},
		{"31337", "0", "0"},
		{"0", "31337", "0"},
		{"2", "31337", "2"},
		{"1", "1", "0"},
		{twoPow256Sub1, "2", twoPow256Sub1}, // -1 mod 2 = -1
		{twoPow256Sub1, "1", "0"},           // -1 mod 1 = 0
		{"12", "115792089237316195423570985008687907853269984665640564039457584007913129639933", "0"}, // 12 mod -3 = 0
		{"13", "115792089237316195423570985008687907853269984665640564039457584007913129639933", "1"}, // 13 mod -3 = 1
		{"115792089237316195423570985008687907853269984665640564039457584007913129639928", "115792089237316195423570985008687907853269984665640564039457584007913129639933", "115792089237316195423570985008687907853269984665640564039457584007913129639934"},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).SMod(x, y)

		if got.Neq(want) {
			t.Errorf("Mod(%s, %s) = %v, want %v", tt.x, tt.y, got.String(), want.String())
		}
	}
}

func TestMulMod(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		m    string
		want string
	}{
		{"0x1", "0x1", "0x2", "0x1"},
		{"0x10", "0x10", "0x7", "0x4"},
		{"0x100", "0x100", "0x17", "0x9"},
		{"0x31337", "0x31337", "0x31338", "0x1"},
		{"0x0", "0x31337", "0x31338", "0x0"},
		{"0x31337", "0x0", "0x31338", "0x0"},
		{"0x2", "0x3", "0x5", "0x1"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0x0"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe", "0x1"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "0xffffffffffffffffffffffffffffffff", "0x0"},
	}

	for _, tt := range tests {
		x := MustFromHex(tt.x)
		y := MustFromHex(tt.y)
		m := MustFromHex(tt.m)
		want := MustFromHex(tt.want)

		got := new(Uint).MulMod(x, y, m)

		if got.Neq(want) {
			t.Errorf(
				"MulMod(%s, %s, %s) = %s, want %s",
				tt.x, tt.y, tt.m, got.String(), want.String(),
			)
		}
	}
}

func TestDivMod(t *testing.T) {
	tests := []struct {
		x       string
		y       string
		wantDiv string
		wantMod string
	}{
		{"1", "1", "1", "0"},
		{"10", "10", "1", "0"},
		{"100", "10", "10", "0"},
		{"31337", "3", "10445", "2"},
		{"31337", "0", "0", "0"},
		{"0", "31337", "0", "0"},
		{"2", "31337", "0", "2"},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		wantDiv := MustFromDecimal(tt.wantDiv)
		wantMod := MustFromDecimal(tt.wantMod)

		gotDiv := new(Uint)
		gotMod := new(Uint)
		gotDiv.DivMod(x, y, gotMod)

		for i := range gotDiv.arr {
			if gotDiv.arr[i] != wantDiv.arr[i] {
				t.Errorf("DivMod(%s, %s) got Div %v, want Div %v", tt.x, tt.y, gotDiv, wantDiv)
				break
			}
		}
		for i := range gotMod.arr {
			if gotMod.arr[i] != wantMod.arr[i] {
				t.Errorf("DivMod(%s, %s) got Mod %v, want Mod %v", tt.x, tt.y, gotMod, wantMod)
				break
			}
		}
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		x    string
		want string
	}{
		{"31337", "115792089237316195423570985008687907853269984665640564039457584007913129608599"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129608599", "31337"},
		{"0", "0"},
		{"2", "115792089237316195423570985008687907853269984665640564039457584007913129639934"},
		{"1", twoPow256Sub1},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		want := MustFromDecimal(tt.want)

		got := new(Uint).Neg(x)

		if got.Neq(want) {
			t.Errorf("Neg(%s) = %v, want %v", tt.x, got.String(), want.String())
		}
	}
}

func TestExp(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "3", "30773171189753"},
		{"31337", "0", "1"},
		{"0", "31337", "0"},
		{"1", "1", "1"},
		{"2", "3", "8"},
		{"2", "64", "18446744073709551616"},
		{"2", "128", "340282366920938463463374607431768211456"},
		{"2", "255", "57896044618658097711785492504343953926634992332820282019728792003956564819968"},
		{"2", "256", "0"}, // overflow
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).Exp(x, y)

		if got.Neq(want) {
			t.Errorf(
				"Exp(%s, %s) = %v, want %v",
				tt.x, tt.y, got.String(), want.String(),
			)
		}
	}
}

func TestExtendSign(t *testing.T) {
	tests := []binOp2Test{
		{"31337", "0", "105"},
		{"31337", "1", "31337"},
		{"1", "1", "1"},
		{"127", "0", "127"},
		{"247", "0", "115792089237316195423570985008687907853269984665640564039457584007913129639927"},
		{"255", "0", twoPow256Sub1},
		{"255", "0", twoPow256Sub1},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		y := MustFromDecimal(tt.y)
		want := MustFromDecimal(tt.want)

		got := new(Uint).ExtendSign(x, y)

		if got.Neq(want) {
			t.Errorf(
				"ExtendSign(%s, %s) = %v, want %v",
				tt.x, tt.y, got.String(), want.String(),
			)
		}
	}
}

func TestExp_LargeExponent(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		exponent string
		expected string
	}{
		{
			name:     "2^129",
			base:     "2",
			exponent: "680564733841876926926749214863536422912",
			expected: "0",
		},
		{
			name:     "2^193",
			base:     "2",
			exponent: "12379400392853802746563808384000000000000000000",
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := MustFromDecimal(tt.base)
			exponent := MustFromDecimal(tt.exponent)
			expected := MustFromDecimal(tt.expected)

			result := new(Uint).Exp(base, exponent)

			if result.Neq(expected) {
				t.Errorf(
					"Test %s failed. Expected %s, got %s",
					tt.name, expected.String(), result.String(),
				)
			}
		})
	}
}
