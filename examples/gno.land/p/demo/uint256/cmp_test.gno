package uint256

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestSign(t *testing.T) {
	tests := []struct {
		input    *Uint
		expected int
	}{
		{
			input:    NewUint(0),
			expected: 0,
		},
		{
			input:    NewUint(1),
			expected: 1,
		},
		{
			input:    NewUint(0x7fffffffffffffff),
			expected: 1,
		},
		{
			input:    NewUint(0x8000000000000000),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			result := tt.input.Sign()
			uassert.Equal(t, result, tt.expected)
		})
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
		{"1", "1", 0},
		{"10", "10", 0},
		{"10", "11", -1},
		{"11", "10", 1},
	}

	for _, tc := range tests {
		x := MustFromDecimal(tc.x)
		y := MustFromDecimal(tc.y)

		uassert.Equal(t, x.Cmp(y), tc.want)
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0", true},
		{"1", false},
		{"10", false},
	}

	for _, tt := range tests {
		x := MustFromDecimal(tt.x)
		uassert.Equal(t, x.IsZero(), tt.want)
	}
}

func TestLtUint64(t *testing.T) {
	tests := []struct {
		x    string
		y    uint64
		want bool
	}{
		{"0", 1, true},
		{"1", 0, false},
		{"10", 10, false},
		{"0xffffffffffffffff", 0, false},
		{"0x10000000000000000", 10000000000000000, false},
	}

	for _, tc := range tests {
		x := parseTestString(t, tc.x)
		uassert.Equal(t, x.LtUint64(tc.y), tc.want)
	}
}

func TestUint_GtUint64(t *testing.T) {
	tests := []struct {
		name string
		z    string
		n    uint64
		want bool
	}{
		{
			name: "z > n",
			z:    "1",
			n:    0,
			want: true,
		},
		{
			name: "z < n",
			z:    "18446744073709551615",
			n:    0xFFFFFFFFFFFFFFFF,
			want: false,
		},
		{
			name: "z == n",
			z:    "18446744073709551615",
			n:    0xFFFFFFFFFFFFFFFF,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := MustFromDecimal(tt.z)
			uassert.Equal(t, z.GtUint64(tt.n), tt.want)
		})
	}
}

func TestSGT(t *testing.T) {
	x := MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	y := MustFromHex("0x0")
	uassert.False(t, x.Sgt(y))

	x = MustFromHex("0x0")
	y = MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	uassert.True(t, x.Sgt(y))
}

func TestEq(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{"0xffffffffffffffff", "18446744073709551615", true},
		{"0x10000000000000000", "18446744073709551616", true},
		{"0", "0", true},
		{twoPow256Sub1, twoPow256Sub1, true},
	}

	for _, tt := range tests {
		x := parseTestString(t, tt.x)

		y, err := FromDecimal(tt.y)
		if err != nil {
			t.Error(err)
			continue
		}

		uassert.Equal(t, x.Eq(y), tt.want)
	}
}

func TestUint_Lte(t *testing.T) {
	tests := []struct {
		z, x string
		want bool
	}{
		{"10", "20", true},
		{"20", "10", false},
		{"10", "10", true},
		{"0", "0", true},
	}

	for _, tt := range tests {
		z, err := FromDecimal(tt.z)
		uassert.NoError(t, err)

		x, err := FromDecimal(tt.x)
		uassert.NoError(t, err)

		uassert.Equal(t, z.Lte(x), tt.want)
	}
}

func TestUint_Gte(t *testing.T) {
	tests := []struct {
		z, x string
		want bool
	}{
		{"20", "10", true},
		{"10", "20", false},
		{"10", "10", true},
		{"0", "0", true},
	}

	for _, tt := range tests {
		z := parseTestString(t, tt.z)
		x := parseTestString(t, tt.x)

		uassert.Equal(t, z.Gte(x), tt.want)
	}
}

func parseTestString(t *testing.T, s string) *Uint {
	t.Helper()
	var x *Uint

	if strings.HasPrefix(s, "0x") {
		x = MustFromHex(s)
	} else {
		x = MustFromDecimal(s)
	}

	return x
}
