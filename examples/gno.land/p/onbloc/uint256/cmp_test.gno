package uint256

import (
	"strings"
	"testing"
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
			if result != tt.expected {
				t.Errorf("Sign() = %d; want %d", result, tt.expected)
			}
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

		got := x.Cmp(y)
		if got != tc.want {
			t.Errorf("Cmp(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
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

		got := x.IsZero()
		if got != tt.want {
			t.Errorf("IsZero(%s) = %v, want %v", tt.x, got, tt.want)
		}
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

		got := x.LtUint64(tc.y)
		if got != tc.want {
			t.Errorf("LtUint64(%s, %d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
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

			if got := z.GtUint64(tt.n); got != tt.want {
				t.Errorf("Uint.GtUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSGT(t *testing.T) {
	x := MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	y := MustFromHex("0x0")
	actual := x.Sgt(y)
	if actual {
		t.Fatalf("Expected %v false", actual)
	}

	x = MustFromHex("0x0")
	y = MustFromHex("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe")
	actual = x.Sgt(y)
	if !actual {
		t.Fatalf("Expected %v true", actual)
	}
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

		got := x.Eq(y)

		if got != tt.want {
			t.Errorf("Eq(%s, %s) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
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
		if err != nil {
			t.Error(err)
			continue
		}
		x, err := FromDecimal(tt.x)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := z.Lte(x); got != tt.want {
			t.Errorf("Uint.Lte(%v, %v) = %v, want %v", tt.z, tt.x, got, tt.want)
		}
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

		if got := z.Gte(x); got != tt.want {
			t.Errorf("Uint.Gte(%v, %v) = %v, want %v", tt.z, tt.x, got, tt.want)
		}
	}
}

func parseTestString(_ *testing.T, s string) *Uint {
	var x *Uint

	if strings.HasPrefix(s, "0x") {
		x = MustFromHex(s)
	} else {
		x = MustFromDecimal(s)
	}

	return x
}
