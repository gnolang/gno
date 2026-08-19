package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
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

// doublingChain emits levels lo..hi of a value-containment "doubling" chain of
// types named <prefix>N: each level embeds the previous one twice by value, the
// classic exponential vector for go/types' validType walk. The caller declares
// the base level <prefix>0, which is what the chain bottoms out in.
func doublingChain(prefix string, lo, hi int) string {
	var b strings.Builder
	for i := lo; i <= hi; i++ {
		fmt.Fprintf(&b, "type %s%d struct{ a, b [0]%s%d }\n", prefix, i, prefix, i-1)
	}
	return b.String()
}

// fanOutSrc builds a standalone package holding a doubling chain of the given
// depth.
func fanOutSrc(depth int) string {
	return "package x\ntype T0 struct{ v int }\n" + doublingChain("T", 1, depth)
}

// doublingPkgSrc builds an importable package whose exported T tops a doubling
// chain of the given depth. Callers pick a depth whose total stays under
// typeExpansionBudget, so the package is accepted on its own and only a further
// cross-package multiplication pushes a dependent over.
func doublingPkgSrc(pkgName string, depth int) string {
	return fmt.Sprintf("package %s\ntype t0 struct{ v int }\n", pkgName) +
		doublingChain("t", 1, depth) +
		fmt.Sprintf("type T struct{ a, b [0]t%d }\n", depth)
}

// makeBoundResolver returns a pkgResolver that parses source from a fixed map,
// treating any other path (unknown/stdlib) as a leaf.
func makeBoundResolver(t *testing.T, fset *token.FileSet, srcs map[string]string) pkgResolver {
	t.Helper()
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
			// Depth alone must not read as fan-out: each level contains the
			// previous one once, so the walk is linear per type.
			"deep linear chain passes",
			linearChainSrc(120),
			false,
		},
		{
			// ... but the budget bounds the AGGREGATE walk, and a linear chain of
			// depth d costs ~d^2 in total (each of the d types costs O(d)), so an
			// extreme depth is rejected on honest arithmetic rather than as a
			// special case: validType really does visit ~4M nodes here. The cap
			// lands near depth 135, orders of magnitude past any real type (measured
			// max depth in stdlibs/examples is single digits).
			"extreme linear depth rejected on aggregate cost",
			linearChainSrc(300),
			true,
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
		{
			// Interface type-set fan-out via multiple (`;`-separated) type elements
			// rather than a union `|`: the generics guard does not reject this shape
			// (no `|`/`~`), so the bound must count both elements and catch it.
			"multi-element interface value fan-out rejected",
			func() string {
				var b strings.Builder
				b.WriteString("package x\ntype I0 interface{ m() }\n")
				for i := 1; i <= 30; i++ {
					fmt.Fprintf(&b, "type I%d interface{ [0]I%d; [1]I%d }\n", i, i-1, i-1)
				}
				return b.String()
			}(),
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

func TestCheckNoUncountableGenerics(t *testing.T) {
	t.Parallel()

	// wantMsg pins the exact rejection phrasing (empty => must be accepted). The
	// message reaches the consensus-hashed tx result, so its wording is frozen
	// once a rejection is committed; assert it verbatim, not just a substring.
	tt := []struct {
		name    string
		src     string
		wantMsg string
	}{
		// go1.17 code that must NOT be rejected.
		{"plain struct passes", "package x\ntype S struct{ a, b int }\n", ""},
		{
			"ordinary interface passes",
			"package x\nimport \"io\"\ntype I interface{ Read([]byte) (int, error); io.Closer }\n",
			"",
		},
		{
			// `|` as bitwise-or in an expression must not be mistaken for a union.
			"bitwise-or expression passes",
			"package x\nfunc f(a, b int) int { return a | b }\n",
			"",
		},
		{
			// `x[i]` array indexing must not be mistaken for generic instantiation.
			"array indexing passes",
			"package x\nfunc f(a []int) int { return a[0] }\n",
			"",
		},

		// go1.18 generics syntax that must be rejected.
		{"generic type declaration rejected", "package x\ntype W[P any] struct{ a P }\n",
			"generic type declarations are not supported"},
		{"generic function rejected", "package x\nfunc F[T any](x T) T { return x }\n",
			"generic functions are not supported"},
		{"generic fan-out (hole #1) rejected", genericFanOutSrc(40),
			"generic type declarations are not supported"},
		{"interface type union rejected", "package x\ntype N interface{ int | string }\n",
			"interface type unions are not supported"},
		{"interface approximation rejected", "package x\ntype N interface{ ~int }\n",
			"interface approximation (~) terms are not supported"},
		{"union fan-out (hole #2) rejected", unionFanOutSrc(40),
			"interface type unions are not supported"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fset, gofs := parseBoundSrc(t, tc.src)
			err := checkNoUncountableGenerics(fset, gofs)
			if tc.wantMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckNoDotImports(t *testing.T) {
	t.Parallel()

	// As with checkNoUncountableGenerics, the rejection message reaches the
	// consensus-hashed tx result, so pin its exact wording.
	tt := []struct {
		name    string
		src     string
		wantMsg string
	}{
		{"no imports passes", "package x\ntype S struct{ a int }\n", ""},
		{"named import passes", "package x\nimport \"io\"\nvar _ io.Reader\n", ""},
		{"aliased import passes", "package x\nimport zz \"io\"\nvar _ zz.Reader\n", ""},
		{"blank import passes", "package x\nimport _ \"io\"\n", ""},
		{
			"dot import rejected",
			"package x\nimport . \"io\"\nvar _ Reader\n",
			"bound.go:2:8: dot imports are not allowed in Gno",
		},
		{
			// Grouped form, and the earliest dot import must be the one reported
			// even when a later import is also a dot import.
			"grouped dot imports report the earliest",
			"package x\nimport (\n\t\"io\"\n\t. \"errors\"\n\t. \"strings\"\n)\n",
			"bound.go:4:2: dot imports are not allowed in Gno",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fset, gofs := parseBoundSrc(t, tc.src)
			err := checkNoDotImports(fset, gofs)
			if tc.wantMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDotImportFanOutRejected covers the dot-import variant of hole #3: a
// dot-imported type is named by a bare identifier, which the bound scores as a
// leaf, so the cross-package expansion goes uncounted and the deploy reaches
// go/types. Both arms run the whole TypeCheckMemPackage path, and the qualified
// arm is the control: identical containment, but reached through pkg.T, so the
// bound sees it and rejects on its own.
func TestDotImportFanOutRejected(t *testing.T) {
	t.Parallel()

	// The dependency is under budget on its own, so it is legitimately
	// deployable; the entry package continues the chain over its T, and its own
	// local chain also stays under budget.
	dep := &std.MemPackage{
		Type: MPUserProd, Name: "dep", Path: "gno.land/p/demo/dep",
		Files: []*std.MemFile{{Name: "dep.gno", Body: doublingPkgSrc("dep", 8)}},
	}
	getter := mockPackageGetter{dep}

	tt := []struct {
		name    string
		head    string
		wantMsg string
	}{
		{
			"qualified",
			"package fan\nimport \"gno.land/p/demo/dep\"\ntype u0 struct{ a, b [0]dep.T }\n",
			"denial-of-service",
		},
		{
			"dot",
			"package fan\nimport . \"gno.land/p/demo/dep\"\ntype u0 struct{ a, b [0]T }\n",
			errDotImports,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mpkg := &std.MemPackage{
				Type: MPUserProd, Name: "fan", Path: "gno.land/p/demo/fan",
				Files: []*std.MemFile{{
					Name: "fan.gno",
					Body: tc.head + doublingChain("u", 1, 2),
				}},
			}
			_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
				Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantMsg)
		})
	}
}

// TestCheckTypeExpansionBoundImports covers hole #3: a value-containment fan-out
// split across an import chain. Each package is under the per-package budget, but
// validType re-expands imported types without memoizing, so the cumulative walk
// doubles per package. The guard must follow the imports and reject the deploy.
func TestCheckTypeExpansionBoundImports(t *testing.T) {
	t.Parallel()

	// p0: a doubling chain whose count (~57k) is legitimately under budget on its
	// own; p1 embeds p0.T four times, pushing the cross-package count over.
	pkgs := map[string]string{
		"gno.land/r/foobar/p0": doublingPkgSrc("p0", 8),
	}

	// p1..p5: each embeds the previous package's T four times.
	for i, prev := 1, "p0"; i <= 5; i++ {
		name := fmt.Sprintf("p%d", i)
		pkgs["gno.land/r/foobar/"+name] = fmt.Sprintf(
			"package %s\nimport \"gno.land/r/foobar/%s\"\ntype T struct{ a, b, c, d [0]%s.T }\n",
			name, prev, prev)
		prev = name
	}

	fset := token.NewFileSet()

	// Deploying p5 must be rejected: the imported chain doubles across packages.
	resolve := makeBoundResolver(t, fset, pkgs)
	err := checkTypeExpansionBoundImports(fset, "gno.land/r/foobar/p5",
		resolve("gno.land/r/foobar/p5"), resolve, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "denial-of-service")

	// A package that imports only a small dependency must NOT be rejected.
	okResolve := makeBoundResolver(t, fset, map[string]string{
		"gno.land/r/foobar/dep": "package dep\ntype T struct{ a, b int }\n",
		"gno.land/r/foobar/u":   "package u\nimport \"gno.land/r/foobar/dep\"\ntype U struct{ a, b, c, d [0]dep.T }\n",
	})
	err = checkTypeExpansionBoundImports(fset, "gno.land/r/foobar/u",
		okResolve("gno.land/r/foobar/u"), okResolve, nil)
	assert.NoError(t, err)
}

// TestCheckTypeExpansionBoundAggregate pins that the budget bounds the TOTAL
// walk, not the largest single type. go/types runs validType once per declared
// type, so many individually-cheap types sharing one chain sum to a walk no
// per-type cap would catch: each type here costs ~450 nodes, comfortably under
// the budget, but enough of them push the package total past it.
func TestCheckTypeExpansionBoundAggregate(t *testing.T) {
	t.Parallel()

	// A depth-6 chain plus n types that each hang one array off its tip, so each
	// costs ~450 and none is individually remarkable.
	src := func(n int) string {
		var b strings.Builder
		b.WriteString(fanOutSrc(6))
		for i := range n {
			fmt.Fprintf(&b, "type A%d struct{ x [0]T6 }\n", i)
		}
		return b.String()
	}

	for _, tc := range []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"under the aggregate budget passes", 20, false},
		{"over the aggregate budget rejected", 60, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fset, gofs := parseBoundSrc(t, src(tc.n))
			err := checkTypeExpansionBound(fset, gofs)
			if !tc.wantErr {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), "denial-of-service")

			// The rejection must be entirely due to the aggregate: no single
			// declaration reaches the budget, so no per-type cap would catch it.
			c := newExpansionChecker("", gofs, nil, nil)
			var worst uint64
			var worstName string
			for _, specs := range c.declsFor("").byName {
				for _, d := range specs {
					if v := satAdd(1, c.cost(d.spec.Type, "", d.imports)); v > worst {
						worst, worstName = v, d.spec.Name.Name
					}
				}
			}
			assert.Less(t, worst, uint64(typeExpansionBudget),
				"costliest single type (%s) must stay under the budget", worstName)
		})
	}
}

// TestExpansionPkgCacheSharing pins that sharing one cache across the nested
// type checks of an importer — which is what makes dependency parsing linear
// rather than quadratic in the import graph — does not change any verdict. In
// particular a package's own (entry) file set must never be served from, or
// leak into, the cache of resolved dependency sources.
func TestExpansionPkgCacheSharing(t *testing.T) {
	t.Parallel()

	pkgs := map[string]string{
		"gno.land/r/foobar/p0": doublingPkgSrc("p0", 8),
	}
	for i, prev := 1, "p0"; i <= 4; i++ {
		name := fmt.Sprintf("p%d", i)
		pkgs["gno.land/r/foobar/"+name] = fmt.Sprintf(
			"package %s\nimport \"gno.land/r/foobar/%s\"\ntype T struct{ a, b [0]%s.T }\n",
			name, prev, prev)
		prev = name
	}

	fset := token.NewFileSet()
	resolve := makeBoundResolver(t, fset, pkgs)
	rejects := func(pkgPath string, cache *expansionPkgCache) bool {
		return checkTypeExpansionBoundImports(fset, pkgPath, resolve(pkgPath), resolve, cache) != nil
	}

	paths := []string{
		"gno.land/r/foobar/p0", "gno.land/r/foobar/p1", "gno.land/r/foobar/p2",
		"gno.land/r/foobar/p3", "gno.land/r/foobar/p4",
	}
	// Baseline: every package checked with its own private cache. The fixture is
	// only meaningful if it produces both verdicts.
	private := map[string]bool{}
	var verdicts []bool
	for _, pp := range paths {
		private[pp] = rejects(pp, nil)
		verdicts = append(verdicts, private[pp])
	}
	require.Contains(t, verdicts, false, "fixture must accept some packages")
	require.Contains(t, verdicts, true, "fixture must reject some packages")

	// Now all of them through one shared cache, in both directions — a poisoned
	// entry would surface whichever order the importer happens to visit them in.
	reversed := slices.Clone(paths)
	slices.Reverse(reversed)
	for _, order := range [][]string{paths, reversed} {
		shared := newExpansionPkgCache()
		for _, pp := range order {
			assert.Equal(t, private[pp], rejects(pp, shared),
				"package %s: shared cache (order %v) changed the verdict", pp, order)
		}
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

// nearBudgetPkgSrc builds a package whose expansion sits just under the budget:
// a doubling chain topped by an exported T, optionally importing another package
// so a closure can be chained together.
func nearBudgetPkgSrc(name, imp string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n", name)
	if imp != "" {
		fmt.Fprintf(&b, "import %q\ntype prev %s.T\n", imp, path.Base(imp))
	}
	b.WriteString("type t0 struct{ v int }\n")
	b.WriteString(doublingChain("t", 1, 8))
	b.WriteString("type T struct{ a, b [0]t8 }\n")
	return b.String()
}

// TestExpansionNodesCoversClosure pins what the chain charges gas for. The
// per-package budget bounds each package, but one transaction re-type-checks
// every transitive dependency, and those dependencies' source bytes were paid for
// by EARLIER transactions — so a tiny package importing a large pre-deployed
// closure would otherwise trigger arbitrarily much unmetered work. The reported
// count must therefore cover the whole closure, not just the entry package.
func TestExpansionNodesCoversClosure(t *testing.T) {
	t.Parallel()

	nodesFor := func(chainLen int) uint64 {
		var pkgs []*std.MemPackage
		for k := range chainLen {
			imp := ""
			if k > 0 {
				imp = fmt.Sprintf("gno.land/p/demo/p%d", k-1)
			}
			name := fmt.Sprintf("p%d", k)
			pkgs = append(pkgs, &std.MemPackage{
				Type: MPUserProd, Name: name, Path: "gno.land/p/demo/" + name,
				Files: []*std.MemFile{{Name: "p.gno", Body: nearBudgetPkgSrc(name, imp)}},
			})
		}
		// A deliberately tiny package importing only the tip of the chain.
		tip := fmt.Sprintf("gno.land/p/demo/p%d", chainLen-1)
		tiny := &std.MemPackage{
			Type: MPUserProd, Name: "tiny", Path: "gno.land/p/demo/tiny",
			Files: []*std.MemFile{{Name: "t.gno", Body: fmt.Sprintf(
				"package tiny\nimport %q\ntype X %s.T\n", tip, path.Base(tip))}},
		}
		getter := mockPackageGetter(pkgs)
		var nodes uint64
		_, err := TypeCheckMemPackage(tiny, TypeCheckOptions{
			Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
			ProdOnly: true, ExpansionNodes: &nodes,
		})
		require.NoError(t, err)
		return nodes
	}

	short, long := nodesFor(4), nodesFor(12)
	t.Logf("closure nodes: 4-deep=%d  12-deep=%d", short, long)

	// The entry package is a few lines, so anything proportional to the closure
	// can only come from the dependencies.
	assert.Greater(t, long, short*2,
		"reported nodes must grow with the dependency closure, else a tiny package "+
			"importing a large pre-deployed chain goes unpriced")
}

// TestExpansionNodesZeroWhenRejected pins that a REJECTED package reports no
// chargeable nodes. Rejecting stops go/types from running, so the count is the
// cost avoided rather than incurred; charging it would price work the guard just
// prevented and would replace the informative rejection with an out-of-gas error.
func TestExpansionNodesZeroWhenRejected(t *testing.T) {
	t.Parallel()

	mpkg := &std.MemPackage{
		Type: MPUserProd, Name: "x", Path: "gno.land/p/demo/x",
		Files: []*std.MemFile{{Name: "x.gno", Body: fanOutSrc(40)}},
	}
	getter := mockPackageGetter{}
	var nodes uint64
	_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
		Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
		ProdOnly: true, ExpansionNodes: &nodes,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "denial-of-service")
	assert.Zero(t, nodes, "a rejected package must not be charged for a walk that never ran")
}
