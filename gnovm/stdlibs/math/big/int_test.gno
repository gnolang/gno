package big_test

import (
	"math/big"
	"strconv"
	"testing"
)

// mustPanicWith runs fn and asserts it panics with exactly the wanted
// message. Messages shared with Go's math/big are pinned so ported code
// that matches on recover() values keeps working.
func mustPanicWith(t *testing.T, label, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s: did not panic, want %q", label, want)
			return
		}
		if s, ok := r.(string); !ok || s != want {
			t.Errorf("%s: panic = %v, want %q", label, r, want)
		}
	}()
	fn()
}

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
	// DivMod with modulus-aliased divisor. Deliberate divergence from Go:
	// Go's DivMod reads y again after writing m, so m==y silently
	// corrupts the result (q=-1 m=0 here). The bridge decodes all
	// arguments before writing any receiver, so every input aliasing is
	// well-defined and mathematically correct.
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
	// Nil and empty inputs decode to canonical zero.
	z = new(big.Int).SetBytes(nil)
	if z.Sign() != 0 || z.String() != "0" {
		t.Errorf("SetBytes(nil): sign=%d str=%q, want 0/\"0\"", z.Sign(), z.String())
	}
	z = new(big.Int).SetBytes([]byte{})
	if z.Sign() != 0 || z.String() != "0" {
		t.Errorf("SetBytes([]): sign=%d str=%q, want 0/\"0\"", z.Sign(), z.String())
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
	mustPanicWith(t, "SetString base=1", "invalid number base 1", func() {
		new(big.Int).SetString("0", 1)
	})
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
	// Abs of zero stays zero.
	z = big.NewInt(0)
	z.Abs(z)
	if z.Sign() != 0 || z.String() != "0" {
		t.Errorf("Abs(0) = %s sign=%d, want 0 sign=0", z.String(), z.Sign())
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

// Single-result helpers (Quo/Rem/Div/Mod) take their own native-bridge
// call path; the pair forms QuoRem/DivMod are tested separately.

func TestQuo(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"7", "3", "2"},
		{"-7", "3", "-2"},
		{"7", "-3", "-2"},
		{"-7", "-3", "2"},
		{"0", "5", "0"},
	}
	for _, tc := range tests {
		got := new(big.Int).Quo(mustParse(t, tc.x), mustParse(t, tc.y)).String()
		if got != tc.want {
			t.Errorf("Quo(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestRem(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"7", "3", "1"},
		{"-7", "3", "-1"},
		{"7", "-3", "1"},
		{"-7", "-3", "-1"},
	}
	for _, tc := range tests {
		got := new(big.Int).Rem(mustParse(t, tc.x), mustParse(t, tc.y)).String()
		if got != tc.want {
			t.Errorf("Rem(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"7", "3", "2"},
		{"-7", "3", "-3"},
		{"7", "-3", "-2"},
		{"-7", "-3", "3"},
	}
	for _, tc := range tests {
		got := new(big.Int).Div(mustParse(t, tc.x), mustParse(t, tc.y)).String()
		if got != tc.want {
			t.Errorf("Div(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestMod(t *testing.T) {
	tests := []struct {
		x, y, want string
	}{
		{"7", "3", "1"},
		{"-7", "3", "2"},
		{"7", "-3", "1"},
		{"-7", "-3", "2"},
	}
	for _, tc := range tests {
		got := new(big.Int).Mod(mustParse(t, tc.x), mustParse(t, tc.y)).String()
		if got != tc.want {
			t.Errorf("Mod(%s, %s) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestDivByZero(t *testing.T) {
	cases := []struct {
		name string
		run  func()
	}{
		{"Quo", func() { new(big.Int).Quo(big.NewInt(5), big.NewInt(0)) }},
		{"Rem", func() { new(big.Int).Rem(big.NewInt(5), big.NewInt(0)) }},
		{"Div", func() { new(big.Int).Div(big.NewInt(5), big.NewInt(0)) }},
		{"Mod", func() { new(big.Int).Mod(big.NewInt(5), big.NewInt(0)) }},
		{"QuoRem", func() {
			new(big.Int).QuoRem(big.NewInt(5), big.NewInt(0), new(big.Int))
		}},
		{"DivMod", func() {
			new(big.Int).DivMod(big.NewInt(5), big.NewInt(0), new(big.Int))
		}},
	}
	for _, tc := range cases {
		mustPanicWith(t, tc.name+" by zero", "division by zero", tc.run)
	}
}

func TestQuoRemDistinctOutputs(t *testing.T) {
	// Passing the same *Int for z and r must panic: the natural
	// last-write-wins order would silently overwrite the quotient with
	// the remainder (which is what Go does in this corner).
	mustPanicWith(t, "QuoRem z==r",
		"big: QuoRem requires distinct *Int for quotient and remainder", func() {
			z := new(big.Int)
			z.QuoRem(big.NewInt(17), big.NewInt(5), z)
		})
}

func TestDivModDistinctOutputs(t *testing.T) {
	mustPanicWith(t, "DivMod z==m",
		"big: DivMod requires distinct *Int for quotient and modulus", func() {
			z := new(big.Int)
			z.DivMod(big.NewInt(17), big.NewInt(5), z)
		})
}

func TestSetStringEdgeCases(t *testing.T) {
	// Pin Go-math/big-stdlib behaviors so a future Go upgrade can't
	// silently shift Gno semantics.
	tests := []struct {
		s    string
		base int
		want string
		ok   bool
	}{
		{"+1", 10, "1", true},      // leading + accepted
		{"-0", 10, "0", true},      // canonical zero
		{"", 10, "", false},        // empty rejected
		{"0x", 0, "", false},       // dangling base prefix
		{"0o755", 0, "493", true},  // octal prefix
		{"5_000", 0, "5000", true}, // digit separator (base 0 only)
		{"5_000", 10, "", false},   // underscore rejected outside base 0
		{"0377", 0, "255", true},   // legacy leading-0 octal
		{"08", 0, "", false},       // 8 is not an octal digit
		{"00", 0, "0", true},       // repeated zeros still base-8 zero
		{"0x_10", 0, "16", true},   // underscore after prefix ok
		{"1__0", 0, "", false},     // doubled underscore rejected
		{"_1", 0, "", false},       // leading underscore rejected
		{"-0X10", 0, "-16", true},  // sign + upper-case prefix
		{"+0o10", 0, "8", true},    // sign + octal prefix
		{"0b", 0, "", false},       // dangling binary prefix
		{"ZZ", 62, "3843", true},   // 'Z' is 61 in base 62
		{"zz", 36, "1295", true},   // 'z' is 35 in base 36
		{"Zz", 62, "3817", true},   // case-sensitive digits above 36
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

func TestTextInvalidBase(t *testing.T) {
	for _, b := range []int{-1, 0, 1, 63, 100} {
		// Unlike SetString, Go's Text message carries no base number.
		mustPanicWith(t, "Text("+strconv.Itoa(b)+")", "invalid base", func() {
			big.NewInt(255).Text(b)
		})
	}
}

func TestNilReceiverText(t *testing.T) {
	var x *big.Int
	if got := x.String(); got != "<nil>" {
		t.Errorf(`nil.String() = %q, want "<nil>"`, got)
	}
	if got := x.Text(16); got != "<nil>" {
		t.Errorf(`nil.Text(16) = %q, want "<nil>"`, got)
	}
	// Like Go, the nil check precedes base validation: no panic even
	// for an invalid base.
	if got := x.Text(1); got != "<nil>" {
		t.Errorf(`nil.Text(1) = %q, want "<nil>"`, got)
	}
}

func TestTextNegativeAndBase62(t *testing.T) {
	tests := []struct {
		v    int64
		base int
		want string
	}{
		{-255, 16, "-ff"},
		{-255, 2, "-11111111"},
		{-255, 62, "-47"},
		{3843, 62, "ZZ"}, // upper-case letters for digit values 36-61
	}
	for _, tc := range tests {
		if got := big.NewInt(tc.v).Text(tc.base); got != tc.want {
			t.Errorf("(%d).Text(%d) = %q, want %q", tc.v, tc.base, got, tc.want)
		}
	}
}

func TestLow64Truncation(t *testing.T) {
	// Go documents Uint64/Int64 as undefined when the value doesn't fit,
	// but the actual behavior is stable (low 64 bits of the magnitude,
	// negated for negative x) and code in the wild relies on it. Pin it.
	wide := mustParse(t, "36893488147419103232") // 2^65
	if got := wide.Uint64(); got != 0 {
		t.Errorf("2^65 Uint64() = %d, want 0", got)
	}
	if got := wide.Int64(); got != 0 {
		t.Errorf("2^65 Int64() = %d, want 0", got)
	}
	nwide := mustParse(t, "-36893488147419103233") // -(2^65+1)
	if got := nwide.Uint64(); got != 1 {
		t.Errorf("-(2^65+1) Uint64() = %d, want 1", got)
	}
	if got := nwide.Int64(); got != -1 {
		t.Errorf("-(2^65+1) Int64() = %d, want -1", got)
	}
}

func TestZeroCanonical(t *testing.T) {
	// Every op that can produce zero must produce the canonical form
	// (Sign == 0, String == "0"). This guards against any future change
	// to the bridge or Gno-side normalization that would create a
	// negative-zero or leading-zero-byte state.
	a := mustParse(t, "100000000000000000000")
	cases := []struct {
		name string
		z    *big.Int
	}{
		{"Sub(x,x)", new(big.Int).Sub(a, a)},
		{"Mul(x,0)", new(big.Int).Mul(a, big.NewInt(0))},
		{"Mul(0,x)", new(big.Int).Mul(big.NewInt(0), a)},
		{"Quo(0,x)", new(big.Int).Quo(big.NewInt(0), a)},
		{"Rem(x,1)", new(big.Int).Rem(a, big.NewInt(1))},
		{"Mod(x,1)", new(big.Int).Mod(a, big.NewInt(1))},
	}
	for _, tc := range cases {
		if tc.z.Sign() != 0 || tc.z.String() != "0" {
			t.Errorf("%s: sign=%d str=%q, want 0/\"0\"",
				tc.name, tc.z.Sign(), tc.z.String())
		}
	}
}
