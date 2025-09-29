package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRealmPath(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input  string
		result bool
	}{
		{"gno.land/r/demo/users", true},
		{"gno.land/r/hello", true},
		{"gno.land/p/demo/users", false},
		{"gno.land/p/hello", false},
		{"gno.land/x", false},
		{"std", false},
	}

	for _, tc := range tt {
		assert.Equal(
			t,
			tc.result,
			IsRealmPath(tc.input),
			"unexpected IsRealmPath(%q) result", tc.input,
		)
	}
}

func TestIsStdlib(t *testing.T) {
	t.Parallel()

	tt := []struct {
		s      string
		result bool
	}{
		{"std", true},
		{"math", true},
		{"very/long/path/with_underscores", true},
		{"gno.land/r/demo/users", false},
		{"gno.land/hello", false},
	}

	for _, tc := range tt {
		assert.Equal(
			t,
			tc.result,
			IsStdlib(tc.s),
			"IsStdlib(%q)", tc.s,
		)
	}
}

func TestIsEphemeralPath(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		expected bool
	}{
		{
			name:     "valid ephemeral path",
			pkgPath:  "gno.land/e/user123/test",
			expected: true,
		},
		{
			name:     "valid ephemeral run path",
			pkgPath:  "gno.land/e/g1user123/run",
			expected: true,
		},
		{
			name:     "valid ephemeral path with subdirectories",
			pkgPath:  "gno.land/e/user123/subdir/test",
			expected: true,
		},
		{
			name:     "realm path should not be ephemeral",
			pkgPath:  "gno.land/r/user123/test",
			expected: false,
		},
		{
			name:     "pure package path should not be ephemeral",
			pkgPath:  "gno.land/p/user123/test",
			expected: false,
		},
		{
			name:     "stdlib path should not be ephemeral",
			pkgPath:  "fmt",
			expected: false,
		},
		{
			name:     "empty path should not be ephemeral",
			pkgPath:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEphemeralPath(tt.pkgPath)
			if result != tt.expected {
				t.Errorf("IsEphemeralPath(%q) = %v, want %v", tt.pkgPath, result, tt.expected)
			}
		})
	}
}

func TestIsGnoRunPath(t *testing.T) {
	validAddr := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	tests := []struct {
		name         string
		pkgPath      string
		expectedAddr string
		expectedOk   bool
	}{
		{
			name:         "valid ephemeral run path",
			pkgPath:      "gno.land/e/" + validAddr + "/run",
			expectedAddr: validAddr,
			expectedOk:   true,
		},
		{
			name:         "old run path should not match",
			pkgPath:      "gno.land/r/" + validAddr + "/run",
			expectedAddr: "",
			expectedOk:   false,
		},
		{
			name:         "ephemeral path without run should not match",
			pkgPath:      "gno.land/e/" + validAddr + "/test",
			expectedAddr: "",
			expectedOk:   false,
		},
		{
			name:         "invalid address format should not match",
			pkgPath:      "gno.land/e/user123/run",
			expectedAddr: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, ok := IsGnoRunPath(tt.pkgPath)
			if ok != tt.expectedOk {
				t.Errorf("IsGnoEphemeralRunPath(%q) ok = %v, want %v", tt.pkgPath, ok, tt.expectedOk)
			}
			if addr != tt.expectedAddr {
				t.Errorf("IsGnoEphemeralRunPath(%q) addr = %v, want %v", tt.pkgPath, addr, tt.expectedAddr)
			}
		})
	}
}

func TestLeftmostX(t *testing.T) {
	t.Parallel()

	// helper functions in this test context.
	name := func(n string) *NameExpr { return &NameExpr{Name: Name(n)} }
	sel := func(x Expr, s string) *SelectorExpr { return &SelectorExpr{X: x, Sel: Name(s)} }
	idx := func(x, i Expr) *IndexExpr { return &IndexExpr{X: x, Index: i} }
	slc := func(x, low, high, max Expr) *SliceExpr {
		return &SliceExpr{X: x, Low: low, High: high, Max: max}
	}
	call := func(fn Expr, args ...Expr) *CallExpr { return &CallExpr{Func: fn, Args: args} }
	star := func(x Expr) *StarExpr { return &StarExpr{X: x} }

	cases := []struct {
		name string
		in   Expr
		want Expr
	}{
		{
			name: "simple selector a.b.c -> a",
			in:   sel(sel(name("a"), "b"), "c"),
			want: name("a"),
		},
		{
			name: "index then selectors a[0].b.c -> a",
			in:   sel(sel(idx(name("a"), name("0")), "b"), "c"),
			want: name("a"),
		},
		{
			name: "slice then selector a[:].b -> a",
			in:   sel(slc(name("a"), nil, nil, nil), "b"),
			want: name("a"),
		},
		{
			name: "call boundary f().b -> f()",
			in:   sel(call(name("f")), "b"),
			want: call(name("f")),
		},
		{
			name: "deref preserved (*a).b -> *a",
			in:   sel(star(name("a")), "b"),
			want: star(name("a")),
		},
		{
			name: "nested index in selector a.b[c.d].e -> a",
			in:   sel(idx(sel(name("a"), "b"), sel(name("c"), "d")), "e"),
			want: name("a"),
		},
		{
			name: "type assert boundary v.(T).field -> v.(T)",
			in:   sel(TypeAssert(name("v"), "T"), "field"),
			want: TypeAssert(name("v"), "T"),
		},
		{
			name: "map index then selector m[k].field -> m",
			in:   sel(idx(name("m"), name("k")), "field"),
			want: name("m"),
		},
		{
			name: "method call boundary ptr.Method().field -> ptr.Method()",
			in:   sel(call(sel(name("ptr"), "Method")), "field"),
			want: call(sel(name("ptr"), "Method")),
		},
	}

	for _, tc := range cases {
		got := LeftmostX(tc.in)
		if gotStr, wantStr := toExprTrace(got), toExprTrace(tc.want); gotStr != wantStr {
			t.Fatalf("%s: leftmost mismatch\n got:  %s\n want: %s", tc.name, gotStr, wantStr)
		}
	}
}
