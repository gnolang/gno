package uint256

import (
	"testing"
)

func TestIsUint64(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0x0", true},
		{"0x1", true},
		{"0x10", true},
		{"0xffffffffffffffff", true},
		{"0x10000000000000000", false},
	}

	for _, tc := range tests {
		x := MustFromHex(tc.x)
		got := x.IsUint64()

		if got != tc.want {
			t.Errorf("IsUint64(%s) = %v, want %v", tc.x, got, tc.want)
		}
	}
}

func TestDec(t *testing.T) {
	testCases := []struct {
		name string
		z    Uint
		want string
	}{
		{
			name: "zero",
			z:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: "0",
		},
		{
			name: "less than 20 digits",
			z:    Uint{arr: [4]uint64{1234567890, 0, 0, 0}},
			want: "1234567890",
		},
		{
			name: "max possible value",
			z:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.z.Dec()
			if result != tc.want {
				t.Errorf("Dec(%v) = %s, want %s", tc.z, result, tc.want)
			}
		})
	}
}
