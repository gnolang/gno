package int256

import (
	"gno.land/p/demo/uint256"
)

import (
	"testing"
)

func TestOr(t *testing.T) {
	tests := []struct {
		name       string
		x, y, want Int
	}{
		{
			name: "all zeroes",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "all ones",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
		},
		{
			name: "mixed",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
		},
		{
			name: "one operand all ones",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New()
			got.Or(&tc.x, &tc.y)

			if got.Neq(&tc.want) {
				t.Errorf("Or(%v, %v) = %v, want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		name       string
		x, y, want Int
	}{
		{
			name: "all zeroes",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "all ones",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
		},
		{
			name: "mixed",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "mixed 2",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "mixed 3",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "one operand zero",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
		},
		{
			name: "one operand all ones",
			x:    Int{abs: &uint256.Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}, neg: false},
			y:    Int{abs: &uint256.Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}}, neg: false},
			want: Int{abs: &uint256.Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}}, neg: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New()
			got.And(&tc.x, &tc.y)

			if got.Neq(&tc.want) {
				t.Errorf("And(%v, %v) = %v, want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestRsh(t *testing.T) {
	tests := []struct {
		x    string
		n    uint
		want string
	}{
		{"1024", 0, "1024"},
		{"1024", 1, "512"},
		{"1024", 2, "256"},
		{"1024", 10, "1"},
		{"1024", 11, "0"},
		{"18446744073709551615", 0, "18446744073709551615"},
		{"18446744073709551615", 1, "9223372036854775807"},
		{"18446744073709551615", 62, "3"},
		{"18446744073709551615", 63, "1"},
		{"18446744073709551615", 64, "0"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", 0, "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", 1, "57896044618658097711785492504343953926634992332820282019728792003956564819967"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", 128, "340282366920938463463374607431768211455"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", 255, "1"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", 256, "0"},
		{"-1024", 0, "-1024"},
		{"-1024", 1, "-512"},
		{"-1024", 2, "-256"},
		{"-1024", 10, "-1"},
		{"-1024", 10, "-1"},
		{"-9223372036854775808", 0, "-9223372036854775808"},
		{"-9223372036854775808", 1, "-4611686018427387904"},
		{"-9223372036854775808", 62, "-2"},
		{"-9223372036854775808", 63, "-1"},
		{"-9223372036854775808", 64, "-1"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 0, "-57896044618658097711785492504343953926634992332820282019728792003956564819968"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 1, "-28948022309329048855892746252171976963317496166410141009864396001978282409984"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 253, "-4"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 254, "-2"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 255, "-1"},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819968", 256, "-1"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Rsh(x, tc.n)

		if got.ToString() != tc.want {
			t.Errorf("Rsh(%s, %d) = %v, want %v", tc.x, tc.n, got.ToString(), tc.want)
		}
	}
}

func TestLsh(t *testing.T) {
	tests := []struct {
		x    string
		n    uint
		want string
	}{
		{"1", 0, "1"},
		{"1", 1, "2"},
		{"1", 2, "4"},
		{"2", 0, "2"},
		{"2", 1, "4"},
		{"2", 2, "8"},
		{"-2", 0, "-2"},
		{"-4", 0, "-4"},
		{"-8", 0, "-8"},
	}

	for _, tc := range tests {
		x, err := FromDecimal(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := New()
		got.Lsh(x, tc.n)

		if got.ToString() != tc.want {
			t.Errorf("Lsh(%s, %d) = %v, want %v", tc.x, tc.n, got.ToString(), tc.want)
		}
	}
}
