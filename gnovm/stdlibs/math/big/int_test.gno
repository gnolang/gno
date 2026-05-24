package big_test

import (
	"math/big"
	"testing"
)

func TestNewInt(t *testing.T) {
	tests := []struct {
		v    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{42, "42"},
		{-42, "-42"},
		{1<<62 + 1, "4611686018427387905"},
		{-1 << 63, "-9223372036854775808"},
	}
	for _, tc := range tests {
		got := big.NewInt(tc.v).String()
		if got != tc.want {
			t.Errorf("NewInt(%d).String() = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestInt64RoundTrip(t *testing.T) {
	tests := []int64{0, 1, -1, 1 << 62, -1 << 62, 1<<63 - 1, -1 << 63}
	for _, v := range tests {
		got := big.NewInt(v).Int64()
		if got != v {
			t.Errorf("NewInt(%d).Int64() = %d", v, got)
		}
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		v    int64
		want int
	}{
		{0, 0},
		{1, 1},
		{-1, -1},
		{1 << 32, 1},
		{-1 << 32, -1},
	}
	for _, tc := range tests {
		got := big.NewInt(tc.v).Sign()
		if got != tc.want {
			t.Errorf("NewInt(%d).Sign() = %d, want %d", tc.v, got, tc.want)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "0", "0"},
		{"1", "2", "3"},
		{"-1", "1", "0"},
		{"-5", "3", "-2"},
		{"100000000000000000000", "1", "100000000000000000001"},
		{"100000000000000000000", "-100000000000000000000", "0"},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		got := new(big.Int).Add(x, y).String()
		if got != tc.want {
			t.Errorf("Add(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "0", "0"},
		{"3", "1", "2"},
		{"1", "3", "-2"},
		{"-5", "-3", "-2"},
		{"100000000000000000000", "1", "99999999999999999999"},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		got := new(big.Int).Sub(x, y).String()
		if got != tc.want {
			t.Errorf("Sub(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"0", "1234", "0"},
		{"1", "0", "0"},
		{"2", "3", "6"},
		{"-2", "3", "-6"},
		{"-2", "-3", "6"},
		// 2^64 * 2^64 = 2^128
		{"18446744073709551616", "18446744073709551616", "340282366920938463463374607431768211456"},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		got := new(big.Int).Mul(x, y).String()
		if got != tc.want {
			t.Errorf("Mul(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestQuoRem(t *testing.T) {
	// Truncated division (Go semantics).
	tests := []struct {
		x, y, wantQ, wantR string
	}{
		{"7", "3", "2", "1"},
		{"-7", "3", "-2", "-1"},
		{"7", "-3", "-2", "1"},
		{"-7", "-3", "2", "-1"},
		{"0", "5", "0", "0"},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(x, y, r)
		if q.String() != tc.wantQ || r.String() != tc.wantR {
			t.Errorf("QuoRem(%s, %s) = (%s, %s), want (%s, %s)",
				tc.x, tc.y, q.String(), r.String(), tc.wantQ, tc.wantR)
		}
	}
}

func TestDivMod(t *testing.T) {
	// Euclidean division (Go's Div/Mod): remainder always in [0, |y|).
	tests := []struct {
		x, y, wantQ, wantM string
	}{
		{"7", "3", "2", "1"},
		{"-7", "3", "-3", "2"},
		{"7", "-3", "-2", "1"},
		{"-7", "-3", "3", "2"},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		q, m := new(big.Int), new(big.Int)
		q.DivMod(x, y, m)
		if q.String() != tc.wantQ || m.String() != tc.wantM {
			t.Errorf("DivMod(%s, %s) = (%s, %s), want (%s, %s)",
				tc.x, tc.y, q.String(), m.String(), tc.wantQ, tc.wantM)
		}
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		x, y string
		want int
	}{
		{"0", "0", 0},
		{"1", "2", -1},
		{"2", "1", 1},
		{"-1", "1", -1},
		{"-1", "-1", 0},
		{"-2", "-1", -1},
		{"100000000000000000000", "100000000000000000000", 0},
		{"100000000000000000000", "99999999999999999999", 1},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		got := x.Cmp(y)
		if got != tc.want {
			t.Errorf("Cmp(%s, %s) = %d, want %d", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestAliasing(t *testing.T) {
	// z.Add(z, z), z.Mul(z, z), etc. must work.
	z := big.NewInt(5)
	z.Add(z, z)
	if z.String() != "10" {
		t.Errorf("z.Add(z, z) starting from 5: got %s, want 10", z.String())
	}
	z = big.NewInt(3)
	z.Mul(z, z)
	if z.String() != "9" {
		t.Errorf("z.Mul(z, z) starting from 3: got %s, want 9", z.String())
	}
	// QuoRem with quotient-aliased dividend, separate remainder.
	q := big.NewInt(17)
	r := new(big.Int)
	q.QuoRem(q, big.NewInt(5), r)
	if q.String() != "3" || r.String() != "2" {
		t.Errorf("QuoRem(17, 5) aliased q=x: q=%s r=%s, want q=3 r=2", q.String(), r.String())
	}
	// DivMod with modulus-aliased divisor.
	d := big.NewInt(-7)
	m := big.NewInt(3)
	d.DivMod(d, m, m)
	if d.String() != "-3" || m.String() != "2" {
		t.Errorf("DivMod(-7, 3) aliased q=x, m=y: q=%s m=%s, want q=-3 m=2", d.String(), m.String())
	}
}

func TestIsInt64Boundaries(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"0", true},
		{"9223372036854775807", true},   // MaxInt64
		{"9223372036854775808", false},  // MaxInt64+1
		{"-9223372036854775808", true},  // MinInt64
		{"-9223372036854775809", false}, // MinInt64-1
	}
	for _, tc := range tests {
		z := mustParse(t, tc.s)
		if got := z.IsInt64(); got != tc.want {
			t.Errorf("IsInt64(%s) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestIsUint64Boundaries(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"0", true},
		{"18446744073709551615", true},  // MaxUint64
		{"18446744073709551616", false}, // MaxUint64+1
		{"-1", false},
	}
	for _, tc := range tests {
		z := mustParse(t, tc.s)
		if got := z.IsUint64(); got != tc.want {
			t.Errorf("IsUint64(%s) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestSetBytesNormalization(t *testing.T) {
	// Leading zeros must be stripped.
	z := new(big.Int).SetBytes([]byte{0, 0, 1, 2, 3})
	got := z.Bytes()
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("SetBytes leading-zeros normalization: got %v, want [1 2 3]", got)
	}
	// All-zero input means zero, never neg=true.
	z = new(big.Int).SetBytes([]byte{0, 0, 0})
	if z.Sign() != 0 || z.String() != "0" {
		t.Errorf("SetBytes all-zero: sign=%d str=%q, want 0/\"0\"", z.Sign(), z.String())
	}
}

func TestCmpAbs(t *testing.T) {
	tests := []struct {
		x, y string
		want int
	}{
		{"5", "3", 1},
		{"-5", "3", 1},
		{"5", "-3", 1},
		{"-5", "-3", 1},
		{"3", "3", 0},
		{"-3", "3", 0},
	}
	for _, tc := range tests {
		x := mustParse(t, tc.x)
		y := mustParse(t, tc.y)
		if got := x.CmpAbs(y); got != tc.want {
			t.Errorf("CmpAbs(%s, %s) = %d, want %d", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestSetStringInvalidBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SetString with base=1 did not panic")
		}
	}()
	new(big.Int).SetString("0", 1)
}

func TestBytes(t *testing.T) {
	z := big.NewInt(0x010203)
	got := z.Bytes()
	if len(got) != 3 || got[0] != 0x01 || got[1] != 0x02 || got[2] != 0x03 {
		t.Errorf("Bytes() = %v, want [1 2 3]", got)
	}
	z2 := new(big.Int).SetBytes(got)
	if z2.Cmp(z) != 0 {
		t.Errorf("SetBytes(Bytes()) round trip failed")
	}
	// Zero
	if len(big.NewInt(0).Bytes()) != 0 {
		t.Errorf("NewInt(0).Bytes() should be empty")
	}
}

func TestBitLen(t *testing.T) {
	tests := []struct {
		v    int64
		want int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{255, 8},
		{256, 9},
	}
	for _, tc := range tests {
		got := big.NewInt(tc.v).BitLen()
		if got != tc.want {
			t.Errorf("NewInt(%d).BitLen() = %d, want %d", tc.v, got, tc.want)
		}
	}
}

func TestSetString(t *testing.T) {
	tests := []struct {
		s    string
		base int
		want string
		ok   bool
	}{
		{"0", 10, "0", true},
		{"123", 10, "123", true},
		{"-456", 10, "-456", true},
		{"ff", 16, "255", true},
		{"0xff", 0, "255", true},
		{"0b1010", 0, "10", true},
		{"100000000000000000000", 10, "100000000000000000000", true},
		{"not a number", 10, "", false},
	}
	for _, tc := range tests {
		z, ok := new(big.Int).SetString(tc.s, tc.base)
		if ok != tc.ok {
			t.Errorf("SetString(%q, %d) ok = %v, want %v", tc.s, tc.base, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if z.String() != tc.want {
			t.Errorf("SetString(%q, %d) = %s, want %s", tc.s, tc.base, z.String(), tc.want)
		}
	}
}

func TestText(t *testing.T) {
	z := big.NewInt(255)
	if got := z.Text(16); got != "ff" {
		t.Errorf("Text(16) of 255 = %q, want %q", got, "ff")
	}
	if got := z.Text(2); got != "11111111" {
		t.Errorf("Text(2) of 255 = %q, want %q", got, "11111111")
	}
}

func TestNegAbs(t *testing.T) {
	z := big.NewInt(-5)
	z.Abs(z)
	if z.String() != "5" {
		t.Errorf("Abs(-5) = %s, want 5", z.String())
	}
	z.Neg(z)
	if z.String() != "-5" {
		t.Errorf("Neg(5) = %s, want -5", z.String())
	}
	// Neg of zero stays zero.
	z = big.NewInt(0)
	z.Neg(z)
	if z.Sign() != 0 || z.String() != "0" {
		t.Errorf("Neg(0) = %s sign=%d, want 0 sign=0", z.String(), z.Sign())
	}
}

func mustParse(t *testing.T, s string) *big.Int {
	t.Helper()
	z, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("failed to parse %q", s)
	}
	return z
}
