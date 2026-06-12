package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseBoundSrc parses Go source into the (fset, []*ast.File) shape that
// checkTypeExpansionBound consumes.
func parseBoundSrc(t *testing.T, src string) (*token.FileSet, []*ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "bound.go", src, parser.SkipObjectResolution)
	require.NoError(t, err)
	return fset, []*ast.File{f}
}

// fanOutSrc builds a value-containment "doubling" chain of the given depth:
// each level embeds the previous one twice by value, the classic exponential
// vector for go/types' validType walk.
func fanOutSrc(depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype T0 struct{ v int }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type T%d struct{ a, b [0]T%d }\n", i, i-1)
	}
	return b.String()
}

// linearChainSrc builds a deep but linear chain: each level contains the
// previous one exactly once, so the walk is linear in depth, not exponential.
func linearChainSrc(depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype T0 struct{ v int }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type T%d struct{ a T%d }\n", i, i-1)
	}
	return b.String()
}

func TestCheckTypeExpansionBound(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			"simple",
			"package x\ntype S struct{ a, b int }\n",
			false,
		},
		{
			"deep linear chain passes",
			linearChainSrc(2000),
			false,
		},
		{
			// Pointers break value containment exactly as validType does, so a
			// deep "doubling" chain through pointers must NOT be rejected.
			"pointer fan-out passes",
			func() string {
				var b strings.Builder
				b.WriteString("package x\ntype T0 struct{ v int }\n")
				for i := 1; i <= 60; i++ {
					fmt.Fprintf(&b, "type T%d struct{ a, b *T%d }\n", i, i-1)
				}
				return b.String()
			}(),
			false,
		},
		{
			// Slices, maps, chans likewise break the chain.
			"slice fan-out passes",
			func() string {
				var b strings.Builder
				b.WriteString("package x\ntype T0 struct{ v int }\n")
				for i := 1; i <= 60; i++ {
					fmt.Fprintf(&b, "type T%d struct{ a, b []T%d }\n", i, i-1)
				}
				return b.String()
			}(),
			false,
		},
		{
			"self-referential via pointer passes",
			"package x\ntype List struct{ v int; next *List }\n",
			false,
		},
		{
			"value fan-out rejected",
			fanOutSrc(30),
			true,
		},
		{
			"array-element value fan-out rejected at function scope",
			"package x\nfunc f() {\n" + strings.TrimPrefix(fanOutSrc(30), "package x\n") + "}\n",
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fset, gofs := parseBoundSrc(t, tc.src)
			err := checkTypeExpansionBound(fset, gofs)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "denial-of-service")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCheckTypeExpansionBoundLinearTime asserts the guard itself is linear: a
// depth-1000 fan-out package (which would make validType visit ~2^1000 nodes)
// is rejected near-instantly because the guard memoizes.
func TestCheckTypeExpansionBoundLinearTime(t *testing.T) {
	t.Parallel()
	fset, gofs := parseBoundSrc(t, fanOutSrc(1000))
	err := checkTypeExpansionBound(fset, gofs)
	require.Error(t, err)
}

func BenchmarkCheckTypeExpansionBound(b *testing.B) {
	for _, depth := range []int{100, 1000, 5000} {
		src := fanOutSrc(depth)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "bound.go", src, parser.SkipObjectResolution)
		if err != nil {
			b.Fatal(err)
		}
		gofs := []*ast.File{f}
		b.Run(fmt.Sprintf("fanout-depth-%d", depth), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = checkTypeExpansionBound(fset, gofs)
			}
		})
	}
}
