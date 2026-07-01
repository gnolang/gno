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

// genericFanOutSrc routes the doubling through a generic type parameter: the
// generic W holds its parameter P twice by value, and each A_n embeds W[A_{n-1}]
// by value, so validType still doubles per level. The doubling lives in the type
// argument, which a naive guard drops when it only costs the base type W.
func genericFanOutSrc(depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype W[P any] struct{ a, b [0]P }\ntype A0 struct{ v int }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type A%d struct{ x W[A%d] }\n", i, i-1)
	}
	return b.String()
}

// unionFanOutSrc routes the doubling through interface type-set unions: each I_n
// unions two array types over I_{n-1}, so validType still doubles per level. Type
// sets are a go1.18 generics feature, so this must be rejected before go/types.
func unionFanOutSrc(depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype I0 interface{ m() }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type I%d interface{ [0]I%d | [1]I%d }\n", i, i-1, i-1)
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

func TestCheckNoGenerics(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name    string
		src     string
		wantErr bool
	}{
		// go1.17 code that must NOT be rejected.
		{"plain struct passes", "package x\ntype S struct{ a, b int }\n", false},
		{
			"ordinary interface passes",
			"package x\nimport \"io\"\ntype I interface{ Read([]byte) (int, error); io.Closer }\n",
			false,
		},
		{
			// `|` as bitwise-or in an expression must not be mistaken for a union.
			"bitwise-or expression passes",
			"package x\nfunc f(a, b int) int { return a | b }\n",
			false,
		},
		{
			// `x[i]` array indexing must not be mistaken for generic instantiation.
			"array indexing passes",
			"package x\nfunc f(a []int) int { return a[0] }\n",
			false,
		},

		// go1.18 generics syntax that must be rejected.
		{"generic type declaration rejected", "package x\ntype W[P any] struct{ a P }\n", true},
		{"generic function rejected", "package x\nfunc F[T any](x T) T { return x }\n", true},
		{"generic fan-out (hole #1) rejected", genericFanOutSrc(40), true},
		{"interface type union rejected", "package x\ntype N interface{ int | string }\n", true},
		{"interface approximation rejected", "package x\ntype N interface{ ~int }\n", true},
		{"union fan-out (hole #2) rejected", unionFanOutSrc(40), true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fset, gofs := parseBoundSrc(t, tc.src)
			err := checkNoGenerics(fset, gofs)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCheckTypeExpansionBoundImports covers hole #3: a value-containment fan-out
// split across an import chain. Each package is under the per-package budget, but
// validType re-expands imported types without memoizing, so the cumulative walk
// doubles per package. The guard must follow the imports and reject the deploy.
func TestCheckTypeExpansionBoundImports(t *testing.T) {
	t.Parallel()

	// p0: a depth-16 doubling chain, cost ~2^16, legitimately under budget.
	pkgs := map[string]string{}
	var p0 strings.Builder
	p0.WriteString("package p0\ntype t0 struct{ v int }\n")
	for i := 1; i <= 15; i++ {
		fmt.Fprintf(&p0, "type t%d struct{ a, b [0]t%d }\n", i, i-1)
	}
	p0.WriteString("type T struct{ a, b [0]t15 }\n")
	pkgs["gno.land/r/foobar/p0"] = p0.String()

	// p1..p5: each embeds the previous package's T four times.
	for i, prev := 1, "p0"; i <= 5; i++ {
		name := fmt.Sprintf("p%d", i)
		pkgs["gno.land/r/foobar/"+name] = fmt.Sprintf(
			"package %s\nimport \"gno.land/r/foobar/%s\"\ntype T struct{ a, b, c, d [0]%s.T }\n",
			name, prev, prev)
		prev = name
	}

	fset := token.NewFileSet()
	// makeResolver returns a pkgResolver that parses source from a fixed map,
	// treating any other path (unknown/stdlib) as a leaf.
	makeResolver := func(srcs map[string]string) pkgResolver {
		return func(pkgPath string) []*ast.File {
			src, ok := srcs[pkgPath]
			if !ok {
				return nil
			}
			f, err := parser.ParseFile(fset, pkgPath+".go", src, parser.SkipObjectResolution)
			require.NoError(t, err)
			return []*ast.File{f}
		}
	}

	// Deploying p5 must be rejected: the imported chain doubles across packages.
	resolve := makeResolver(pkgs)
	err := checkTypeExpansionBoundImports(fset, "gno.land/r/foobar/p5",
		resolve("gno.land/r/foobar/p5"), resolve)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "denial-of-service")

	// A package that imports only a small dependency must NOT be rejected.
	okResolve := makeResolver(map[string]string{
		"gno.land/r/foobar/dep": "package dep\ntype T struct{ a, b int }\n",
		"gno.land/r/foobar/u":   "package u\nimport \"gno.land/r/foobar/dep\"\ntype U struct{ a, b, c, d [0]dep.T }\n",
	})
	err = checkTypeExpansionBoundImports(fset, "gno.land/r/foobar/u",
		okResolve("gno.land/r/foobar/u"), okResolve)
	assert.NoError(t, err)
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
