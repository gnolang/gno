package gnolang

import (
	"context"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseCostSrc parses Go source into the (fset, []*ast.File) shape that
// typeExpansionCost consumes.
func parseCostSrc(t *testing.T, src string) (*token.FileSet, []*ast.File) {
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
func doublingChain(prefix string, depth int) string {
	var b strings.Builder
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type %s%d struct{ a, b [0]%s%d }\n", prefix, i, prefix, i-1)
	}
	return b.String()
}

// chainSrc builds a standalone package whose levels each reference the previous
// one twice through elem — "[0]" for value containment (the exponential vector),
// "*" or "[]" for the forms validType does not follow.
func chainSrc(elem string, depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype T0 struct{ v int }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type T%d struct{ a, b %sT%d }\n", i, elem, i-1)
	}
	return b.String()
}

// fanOutSrc is the value-containment chain: the shape that makes validType
// exponential.
func fanOutSrc(depth int) string { return chainSrc("[0]", depth) }

// doublingPkgSrc builds an importable package whose exported T tops a doubling
// chain of the given depth. Callers pick a depth whose own cost is modest, so that
// what a dependent pays is dominated by the cross-package multiplication rather
// than by this package. imp, when non-empty, is imported and referenced, so a
// chain of these forms an import chain.
func doublingPkgSrc(pkgName string, depth int, imp string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n", pkgName)
	if imp != "" {
		fmt.Fprintf(&b, "import %q\ntype prev %s.T\n", imp, path.Base(imp))
	}
	b.WriteString("type t0 struct{ v int }\n")
	b.WriteString(doublingChain("t", depth))
	fmt.Fprintf(&b, "type T struct{ a, b [0]t%d }\n", depth)
	return b.String()
}

// makeCostResolver returns a pkgResolver that parses source from a fixed map,
// treating any other path (unknown/stdlib) as a leaf.
func makeCostResolver(t *testing.T, fset *token.FileSet, srcs map[string]string) pkgResolver {
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
func unionFanOutSrc(depth int) string { return ifaceChainSrc("|", depth) }

// ifaceChainSrc builds an interface type-set chain whose terms are separated by
// sep: "|" is a union (a generics construct cost() cannot count, so rejected),
// ";" is ordinary containment the cost model must count itself.
func ifaceChainSrc(sep string, depth int) string {
	var b strings.Builder
	b.WriteString("package x\ntype I0 interface{ m() }\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type I%d interface{ [0]I%d %s [1]I%d }\n", i, i-1, sep, i-1)
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

// costlyThreshold separates shapes whose validType walk is astronomically
// expensive from ordinary ones. It is not a limit — nothing rejects a package for
// exceeding it — just a value no benign type graph comes near (the largest real
// package measures 181 nodes, see TestHonestTypeExpansionUnderBudget) and every
// fan-out shape blows past.
const costlyThreshold = 1_000_000

// TestTypeExpansionCost pins which containment edges the cost model follows.
// Everything downstream is pricing, so an edge missed here is CPU the walk spends
// for free, and an edge invented here is gas honest code pays for nothing.
func TestTypeExpansionCost(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		src      string
		wantHuge bool
	}{
		{
			"simple",
			"package x\ntype S struct{ a, b int }\n",
			false,
		},
		{
			// Depth alone must not read as fan-out: each level contains the
			// previous one once, so the walk is linear per type.
			"deep linear chain is cheap",
			linearChainSrc(900),
			false,
		},
		{
			// ... but the charge is the AGGREGATE walk, and a linear chain of depth d
			// costs ~d^2 in total (each of the d types costs O(d)), so extreme depth
			// gets expensive on honest arithmetic rather than as a special case:
			// validType really does visit ~4M nodes here. Real types are single-digit
			// deep, so this is orders of magnitude past anything honest.
			"extreme linear depth is expensive on aggregate cost",
			linearChainSrc(2000),
			true,
		},
		{
			// Pointers break value containment exactly as validType does, so a deep
			// "doubling" chain through pointers costs nothing to walk.
			"pointer fan-out is cheap",
			chainSrc("*", 60),
			false,
		},
		{
			// Slices, maps, chans likewise break the chain.
			"slice fan-out is cheap",
			chainSrc("[]", 60),
			false,
		},
		{
			"self-referential via pointer is cheap",
			"package x\ntype List struct{ v int; next *List }\n",
			false,
		},
		{
			"value fan-out is astronomically expensive",
			fanOutSrc(30),
			true,
		},
		{
			"array-element value fan-out counted at function scope",
			"package x\nfunc f() {\n" + strings.TrimPrefix(fanOutSrc(30), "package x\n") + "}\n",
			true,
		},
		{
			// Interface type-set fan-out via multiple (`;`-separated) type elements
			// rather than a union `|`: the generics guard does not reject this shape
			// (no `|`/`~`), so the cost model must count both elements itself.
			"multi-element interface value fan-out counted",
			ifaceChainSrc(";", 30),
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, gofs := parseCostSrc(t, tc.src)
			cost := typeExpansionCost("", gofs, nil, nil)
			if tc.wantHuge {
				assert.Greater(t, cost, uint64(costlyThreshold),
					"this shape drives validType exponential; the cost model must see it")
			} else {
				assert.LessOrEqual(t, cost, uint64(costlyThreshold),
					"this shape is cheap for validType; charging for it would overprice honest code")
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
			fset, gofs := parseCostSrc(t, tc.src)
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
			fset, gofs := parseCostSrc(t, tc.src)
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

	// The dependency is cheap enough to be legitimately deployable on its own; the
	// entry package continues the chain over its T.
	dep := &std.MemPackage{
		Type: MPUserProd, Name: "dep", Path: "gno.land/p/demo/dep",
		Files: []*std.MemFile{{Name: "dep.gno", Body: doublingPkgSrc("dep", 12, "")}},
	}
	getter := mockPackageGetter{dep}

	tt := []struct {
		name    string
		head    string
		wantMsg string
	}{
		{
			// The control: identical containment, but reached through dep.T, so the
			// cost model crosses the import and the deploy is billed for the whole
			// cross-package walk — more than this budget covers.
			"qualified",
			"package fan\nimport \"gno.land/p/demo/dep\"\ntype u0 struct{ a, b [0]dep.T }\n",
			"out of gas",
		},
		{
			// Written as a bare T the same walk is invisible to the cost model, so it
			// would be walked for free. It must not reach go/types at all.
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
					Body: tc.head + doublingChain("u", 9),
				}},
			}
			err := func() (err error) {
				// The chain's meter panics on out-of-gas; that panic is the abort.
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("%v", r)
					}
				}()
				_, e := TypeCheckMemPackage(mpkg, TypeCheckOptions{
					Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
					GasMeter: newRecordingGasMeter(1e7),
				})
				return e
			}()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantMsg)
		})
	}
}

// TestTypeExpansionCostImports covers hole #3: a value-containment fan-out split
// across an import chain. Each package's own source is trivial, but validType
// re-expands imported types without memoizing, so the real walk multiplies at
// every link. The cost model has to follow the imports, or every link past the
// first is walked for free.
func TestTypeExpansionCostImports(t *testing.T) {
	t.Parallel()

	// p0: a doubling chain costing ~57k on its own. Each pN after it declares one
	// three-line type, so its own bytes are trivial while its walk is 4x the
	// previous package's.
	pkgs := map[string]string{
		"gno.land/r/foobar/p0": doublingPkgSrc("p0", 12, ""),
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
	resolve := makeCostResolver(t, fset, pkgs)
	costOf := func(pkgPath string) uint64 {
		return typeExpansionCost(pkgPath, resolve(pkgPath), resolve, nil)
	}

	// Every link at least doubles the one before it, so the price tracks the walk
	// down the whole chain rather than stopping at the first import.
	prev := costOf("gno.land/r/foobar/p0")
	for i := 1; i <= 5; i++ {
		pp := fmt.Sprintf("gno.land/r/foobar/p%d", i)
		cost := costOf(pp)
		assert.GreaterOrEqual(t, cost, 2*prev,
			"p%d costs %d against p%d's %d: the import edge is not being followed",
			i, cost, i-1, prev)
		prev = cost
	}
	assert.Greater(t, prev, uint64(costlyThreshold),
		"five links of cross-package doubling must end up astronomically expensive")

	// And the counterfactual: with imports scored as leaves — what an entry-package
	// guard alone would see — p5 is three lines of source and costs almost nothing.
	// That gap is the whole reason the resolver exists.
	p5 := "gno.land/r/foobar/p5"
	asLeaf := typeExpansionCost(p5, resolve(p5), nil, nil)
	assert.Less(t, asLeaf, uint64(1_000))
	assert.Greater(t, prev, 1_000*asLeaf,
		"not following imports under-prices p5 by %dx", prev/asLeaf)

	// Following imports must not make an ordinary small dependency expensive.
	okResolve := makeCostResolver(t, fset, map[string]string{
		"gno.land/r/foobar/dep": "package dep\ntype T struct{ a, b int }\n",
		"gno.land/r/foobar/u":   "package u\nimport \"gno.land/r/foobar/dep\"\ntype U struct{ a, b, c, d [0]dep.T }\n",
	})
	assert.Less(t, typeExpansionCost("gno.land/r/foobar/u",
		okResolve("gno.land/r/foobar/u"), okResolve, nil), uint64(100))
}

// TestTypeExpansionCostAggregate pins that the price is the TOTAL walk, not the
// largest single type. go/types runs validType once per declared type, so many
// individually-cheap types sharing one chain sum to a walk that pricing the
// largest type alone would hand over almost for free.
func TestTypeExpansionCostAggregate(t *testing.T) {
	t.Parallel()

	// A depth-10 chain plus n types that each hang one array off its tip, so each
	// costs ~7166 and none is individually remarkable.
	src := func(n int) string {
		var b strings.Builder
		b.WriteString(fanOutSrc(10))
		for i := range n {
			fmt.Fprintf(&b, "type A%d struct{ x [0]T10 }\n", i)
		}
		return b.String()
	}

	// The total must scale with the number of declarations, and each declaration
	// must stay unremarkable on its own — otherwise this fixture would not
	// distinguish summing from taking the max.
	_, gofs100 := parseCostSrc(t, src(100))
	_, gofs200 := parseCostSrc(t, src(200))
	cost100 := typeExpansionCost("", gofs100, nil, nil)
	cost200 := typeExpansionCost("", gofs200, nil, nil)
	assert.Greater(t, cost200, cost100+90*(cost100/100),
		"100 more declarations of the same shape must add ~100x one declaration; "+
			"the charge is the sum over declarations, not the costliest one")

	c := newExpansionChecker("", gofs200, nil, nil)
	var worst uint64
	var worstName string
	for _, specs := range c.declsFor("").byName {
		for _, d := range specs {
			if v := satAdd(1, c.cost(d.spec.Type, "", d.imports)); v > worst {
				worst, worstName = v, d.spec.Name.Name
			}
		}
	}
	assert.Less(t, worst*10, cost200,
		"costliest single type (%s, %d) must be a small fraction of the total (%d)",
		worstName, worst, cost200)
}

// TestExpansionPkgCacheSharing pins that sharing one cache across the nested
// type checks of an importer — which is what makes dependency parsing linear
// rather than quadratic in the import graph — does not change any price. In
// particular a package's own (entry) file set must never be served from, or
// leak into, the cache of resolved dependency sources: it is seeded with
// whichever file set its caller is checking, so standing in for a dependency
// would mis-price that dependency.
func TestExpansionPkgCacheSharing(t *testing.T) {
	t.Parallel()

	pkgs := map[string]string{
		"gno.land/r/foobar/p0": doublingPkgSrc("p0", 13, ""),
	}
	for i, prev := 1, "p0"; i <= 4; i++ {
		name := fmt.Sprintf("p%d", i)
		pkgs["gno.land/r/foobar/"+name] = fmt.Sprintf(
			"package %s\nimport \"gno.land/r/foobar/%s\"\ntype T struct{ a, b [0]%s.T }\n",
			name, prev, prev)
		prev = name
	}

	fset := token.NewFileSet()
	resolve := makeCostResolver(t, fset, pkgs)
	costOf := func(pkgPath string, cache *expansionPkgCache) uint64 {
		return typeExpansionCost(pkgPath, resolve(pkgPath), resolve, cache)
	}

	paths := []string{
		"gno.land/r/foobar/p0", "gno.land/r/foobar/p1", "gno.land/r/foobar/p2",
		"gno.land/r/foobar/p3", "gno.land/r/foobar/p4",
	}
	// Baseline: every package priced with its own private cache. The fixture is
	// only meaningful if the packages price differently from one another.
	private := map[string]uint64{}
	for _, pp := range paths {
		private[pp] = costOf(pp, nil)
	}
	require.NotEqual(t, private[paths[0]], private[paths[len(paths)-1]],
		"fixture must produce distinct prices, or cache poisoning would be invisible")

	// Now all of them through one shared cache, in both directions — a poisoned
	// entry would surface whichever order the importer happens to visit them in.
	reversed := slices.Clone(paths)
	slices.Reverse(reversed)
	for _, order := range [][]string{paths, reversed} {
		shared := newExpansionPkgCache()
		for _, pp := range order {
			assert.Equal(t, private[pp], costOf(pp, shared),
				"package %s: shared cache (order %v) changed the price", pp, order)
		}
	}
}

// TestTypeExpansionCostLinearTime asserts the cost model itself is linear: a
// depth-1000 fan-out package, which would make validType visit ~2^1000 nodes, is
// priced near-instantly because the model memoizes what validType does not. The
// count saturates rather than wrapping, and expansionGas turns that into a charge
// nobody can afford instead of a negative one.
func TestTypeExpansionCostLinearTime(t *testing.T) {
	t.Parallel()
	_, gofs := parseCostSrc(t, fanOutSrc(1000))
	cost := typeExpansionCost("", gofs, nil, nil)
	assert.Equal(t, uint64(math.MaxUint64), cost)
	assert.Equal(t, int64(math.MaxInt64), expansionGas(cost))
}

// TestExpansionGas pins the clamp: converting a saturated count straight to int64
// yields -1, and a negative charge would refund gas for the worst package a
// sender could submit.
func TestExpansionGas(t *testing.T) {
	t.Parallel()
	assert.Equal(t, int64(0), expansionGas(0))
	assert.Equal(t, int64(typeExpansionGasPerNode), expansionGas(1))
	assert.Equal(t, int64(math.MaxInt64), expansionGas(math.MaxUint64))
	assert.Equal(t, int64(math.MaxInt64), expansionGas(math.MaxUint64/2))
	// The largest count that still converts exactly.
	exact := uint64(math.MaxInt64) / typeExpansionGasPerNode
	assert.Equal(t, int64(exact)*typeExpansionGasPerNode, expansionGas(exact))
	assert.Equal(t, int64(math.MaxInt64), expansionGas(exact+1))
}

// BenchmarkValidTypeWalk measures the thing typeExpansionGasPerNode prices: the
// wall time go/types spends per node of the validType walk. It reports ns/node,
// which is the rate the constant is derived from — see typeExpansionGasPerNode for
// the derivation and the host-calibration step, which this benchmark does NOT do.
//
// Each iteration type-checks a whole doubling chain, so an op is ~0.5s at depth 20
// and doubles per level. Run it as
//
//	go test ./gnovm/pkg/gnolang -run '^$' -bench ValidTypeWalk -benchtime 1x
//
// The rate climbs with depth (the working set outgrows cache), and the DoS case is
// the large-working-set end, so read the deepest figure, not the shallowest.
func BenchmarkValidTypeWalk(b *testing.B) {
	for _, depth := range []int{18, 20, 22} {
		src := fanOutSrc(depth)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "x.go", src, 0)
		if err != nil {
			b.Fatal(err)
		}
		gofs := []*ast.File{f}
		nodes := typeExpansionCost("", gofs, nil, nil)

		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			for range b.N {
				conf := types.Config{Importer: importer.Default(), Error: func(error) {}}
				_, _ = conf.Check("x", fset, gofs, nil)
			}
			b.ReportMetric(
				float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(nodes),
				"ns/node")
			b.ReportMetric(float64(nodes), "nodes")
		})
	}
}

func BenchmarkTypeExpansionCost(b *testing.B) {
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
				_ = typeExpansionCost("", gofs, nil, nil)
			}
		})
	}
}

// recordingGasMeter wraps a real gas meter so a test can see each individual
// charge, not just the total. Charges past the limit panic exactly as the chain's
// meter does, which is what aborts a type check part-way down an import chain.
type recordingGasMeter struct {
	store.GasMeter
	charges []int64
}

func newRecordingGasMeter(limit int64) *recordingGasMeter {
	return &recordingGasMeter{GasMeter: store.NewGasMeter(limit)}
}

func (m *recordingGasMeter) ConsumeGas(amount store.Gas, descriptor string) {
	m.charges = append(m.charges, amount)
	m.GasMeter.ConsumeGas(amount, descriptor)
}

// TestExpansionChargedPerPackage pins that every transitive dependency is
// charged individually, and that an out-of-gas aborts the walk part-way down the
// import chain, not after every package has been walked. See
// typeExpansionGasPerNode.
func TestExpansionChargedPerPackage(t *testing.T) {
	t.Parallel()

	// run type-checks a tiny package importing a chainLen-deep dependency chain. limit
	// bounds the gas available, so a low limit aborts partway through.
	run := func(chainLen int, limit int64) (charges []int64, aborted bool) {
		var pkgs []*std.MemPackage
		for k := range chainLen {
			imp := ""
			if k > 0 {
				imp = fmt.Sprintf("gno.land/p/demo/p%d", k-1)
			}
			name := fmt.Sprintf("p%d", k)
			pkgs = append(pkgs, &std.MemPackage{
				Type: MPUserProd, Name: name, Path: "gno.land/p/demo/" + name,
				Files: []*std.MemFile{{Name: "p.gno", Body: doublingPkgSrc(name, 8, imp)}},
			})
		}
		tip := fmt.Sprintf("gno.land/p/demo/p%d", chainLen-1)
		tiny := &std.MemPackage{
			Type: MPUserProd, Name: "tiny", Path: "gno.land/p/demo/tiny",
			Files: []*std.MemFile{{Name: "t.gno", Body: fmt.Sprintf(
				"package tiny\nimport %q\ntype X %s.T\n", tip, path.Base(tip))}},
		}
		getter := mockPackageGetter(pkgs)
		meter := newRecordingGasMeter(limit)

		defer func() {
			charges, aborted = meter.charges, recover() != nil
		}()
		_, err := TypeCheckMemPackage(tiny, TypeCheckOptions{
			Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed, ProdOnly: true,
			GasMeter: meter,
		})
		require.NoError(t, err)
		return meter.charges, false
	}

	// Ample gas: every package among the dependencies is charged, not just the entry.
	charges, _ := run(6, 1e9)
	assert.Len(t, charges, 8, "one charge per package type-checked, dependencies included")
	for i, n := range charges {
		assert.NotZero(t, n, "charge %d must be non-zero", i)
	}

	// Gas for roughly three packages: the out-of-gas must abort part-way through the dependencies,
	// leaving most of the 22 packages unwalked.
	charges, aborted := run(20, 3*charges[1])
	assert.True(t, aborted, "out-of-gas must propagate out of go/types")
	assert.Less(t, len(charges), 6,
		"the abort must stop the walk part-way through the dependencies, not after all 22 packages")
}

// TestExpansionNotChargedWhenRejected pins the ordering of the three guards: the
// two that reject on syntax run BEFORE the count is computed, so a package they
// turn away is not billed for a walk that never ran — and, more to the point, the
// cost model is never asked to price a package whose shapes it cannot model.
func TestExpansionNotChargedWhenRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, body, wantMsg string
	}{
		{
			"dot import",
			"package x\nimport . \"std\"\ntype T struct{ v int }\n",
			errDotImports,
		},
		{
			"type parameter",
			"package x\ntype T[A any] struct{ v A }\n",
			"generic type declarations are not supported",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mpkg := &std.MemPackage{
				Type: MPUserProd, Name: "x", Path: "gno.land/p/demo/x",
				Files: []*std.MemFile{{Name: "x.gno", Body: tc.body}},
			}
			getter := mockPackageGetter{}
			meter := newRecordingGasMeter(1e9)
			_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
				Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
				ProdOnly: true, GasMeter: meter,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantMsg)
			assert.Empty(t, meter.charges,
				"a package rejected before the count is computed must not be charged")
		})
	}
}

// hangChildEnv makes the subprocess below run the doomed walk instead of the test.
const hangChildEnv = "GNO_TEST_VALIDTYPE_HANG_CHILD"

// TestValidTypeWalkIsExponential asserts the premise of this whole guard: that
// go/types, left to itself, does NOT finish a 30-level doubling chain. Every other
// test here checks that the charge prices such a package; none of them show that
// the package is dangerous in the first place, so without this one a reader cannot
// tell the vulnerability is real rather than theoretical.
//
// Two halves, same input:
//
//   - Unguarded, in a SUBPROCESS: go/types is handed the chain directly and must
//     still be running when the deadline fires. A subprocess because the walk
//     cannot be cancelled — ~2^31 node visits is minutes of CPU — so it has to be
//     killed rather than left burning a core for the rest of the test binary's life.
//   - Guarded, in-process: the same source through TypeCheckMemPackage with a gas
//     meter returns an out-of-gas error. That this half returns AT ALL is the
//     assertion; if the charge did not land before go/types, this test would hang.
//
// If the first half ever fails — the child finishing on its own — validType has
// probably been memoized upstream (golang/go#65711). That would not make the charge
// wrong, but it would invalidate the reasoning on typeExpansionGasPerNode, so
// re-derive it rather than deleting this test.
func TestValidTypeWalkIsExponential(t *testing.T) {
	const depth = 30

	if os.Getenv(hangChildEnv) == "1" {
		// Child: the unguarded walk. Expected to be killed, never to return.
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "x.go", fanOutSrc(depth), 0)
		require.NoError(t, err)
		conf := types.Config{Importer: importer.Default(), Error: func(error) {}}
		_, _ = conf.Check("x", fset, []*ast.File{f}, nil)
		return
	}
	if testing.Short() {
		t.Skip("spawns a subprocess and waits out a deadline")
	}
	t.Parallel()

	// The count is deterministic, so state it exactly: ~2.1e9 nodes for 31 lines of
	// source. At the ~30ns/node BenchmarkValidTypeWalk measures, that is minutes.
	_, gofs := parseCostSrc(t, fanOutSrc(depth))
	nodes := typeExpansionCost("", gofs, nil, nil)
	require.Greater(t, nodes, uint64(1e9), "fixture no longer produces a huge walk")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0],
		"-test.run=^TestValidTypeWalkIsExponential$", "-test.timeout=0")
	cmd.Env = append(os.Environ(), hangChildEnv+"=1")
	err := cmd.Run()

	require.Error(t, err,
		"go/types finished a depth-%d doubling chain (%d nodes) inside the deadline; "+
			"validType may have been memoized upstream (golang/go#65711), which "+
			"invalidates the rate derivation on typeExpansionGasPerNode", depth, nodes)
	require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded,
		"child failed for some reason other than being killed: %v", err)

	// Same source, now charged for. Returning at all is the point.
	mpkg := &std.MemPackage{
		Type: MPUserProd, Name: "x", Path: "gno.land/p/demo/x",
		Files: []*std.MemFile{{Name: "x.gno", Body: fanOutSrc(depth)}},
	}
	getter := mockPackageGetter{}
	oog := func() (msg string) {
		defer func() {
			if r := recover(); r != nil {
				msg = fmt.Sprintf("%v", r)
			}
		}()
		_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
			Getter: getter, TestGetter: getter, Mode: TCLatestRelaxed,
			ProdOnly: true, GasMeter: newRecordingGasMeter(1e7),
		})
		if err != nil {
			return err.Error()
		}
		return ""
	}()
	assert.Contains(t, oog, "out of gas",
		"the charge must abort the deploy before go/types walks the chain")
}

// chainOn appends a doubling chain of the given depth rooted at base, so a small
// difference in base's score is amplified 2^depth times.
func chainOn(base string, depth int) string {
	var b strings.Builder
	for i := 1; i <= depth; i++ {
		prev := base
		if i > 1 {
			prev = fmt.Sprintf("Z%d", i-1)
		}
		fmt.Fprintf(&b, "type Z%d struct{ p, q %s }\n", i, prev)
	}
	return b.String()
}

// TestTypeExpansionCostIsDeterministic pins that the charge is identical on every
// run for every shape that makes cost() take its one order-sensitive branch.
//
// That branch is the containment-cycle truncation in namedCost: it returns 1 for a
// member it is already visiting and correctly does not memoize that, but every
// ancestor on the cycle memoizes a value DERIVED from the truncation. Which member
// truncates depends on which root is walked first, so an unsorted walk priced the
// same source differently per run — 3076 or 6136 nodes, 306k gas apart. The count
// is charged as gas, so nodes that disagree fork: ABCIResult.Error is hashed into
// LastResultsHash and runTx charges GasConsumedToLimit() to the BlockGasMeter.
//
// Cycles are the whole risk surface here. Every other branch of cost() is a pure
// function of its key, and satAdd/satMul saturate to a fixed point, so neither can
// depend on walk order.
//
// Two things this table is shaped to say out loud:
//
//   - The cases only reachable through INVALID input are the ones that diverged.
//     go/types rejects every one of them a moment later, but the charge lands
//     first, so a determinism suite built only from valid fixtures — the whole
//     examples corpus, say — is silent on this entire class.
//   - "symmetric 2-cycle" is a control: it is stable even unsorted, because both
//     orders truncate an equal-weight member. It is why a first attempt at this
//     test can pass while the bug is live.
//
// The loop is the test: a single call cannot observe map order.
//
// Verified by reverting the sort in declsFor and re-running. Four cases catch it,
// three are controls that cannot:
//
//	2-cycle, chain on one member       3076 / 6136        detects
//	2-cycle, chains on both members     844 / 1564        detects
//	cycle across an import boundary     766 / 1510        detects
//	3-cycle with a chain           822 / 1578 / 3090      detects (three-way)
//	symmetric 2-cycle                        32           control, stable
//	self-reference by value                 751           control, stable
//	acyclic control                        2537           control, stable
//
// If you add a case, check it against a reverted sort. A case that cannot fail is
// documentation, not a test — keep it only if it is labelled a control here.
func TestTypeExpansionCostIsDeterministic(t *testing.T) {
	t.Parallel()

	const cyc = "type A struct{ x, y B }\ntype B struct{ x, y A }\n"

	tt := []struct {
		name  string
		pkgs  map[string]string // entry is "p"; extra packages are resolved
		entry string
	}{
		{
			// Control: equal weight either side, so both orders agree. Stable even
			// with the bug present.
			name: "symmetric 2-cycle",
			pkgs: map[string]string{"p": "package p\n" + cyc},
		},
		{
			// The reported diverger: 8 doubling levels on one member only.
			name: "2-cycle, chain on one member",
			pkgs: map[string]string{"p": "package p\n" + cyc + chainOn("A", 8)},
		},
		{
			// Both members carry weight, at different depths.
			name: "2-cycle, chains on both members",
			pkgs: map[string]string{"p": "package p\n" + cyc +
				chainOn("A", 6) + "type Y1 struct{ p, q B }\ntype Y2 struct{ p, q Y1 }\n"},
		},
		{
			name: "3-cycle with a chain",
			pkgs: map[string]string{"p": "package p\n" +
				"type A struct{ x, y B }\ntype B struct{ x, y C }\ntype C struct{ x, y A }\n" +
				chainOn("B", 6)},
		},
		{
			name: "self-reference by value",
			pkgs: map[string]string{"p": "package p\ntype S struct{ v int; next S }\n" +
				chainOn("S", 6)},
		},
		{
			// Cross-package cycle. visiting is keyed by (package, name), so the
			// truncation can land in either package — a path nothing else covers.
			name:  "cycle across an import boundary",
			entry: "p",
			pkgs: map[string]string{
				"p": "package p\nimport \"q\"\ntype A struct{ x, y q.B }\n" + chainOn("A", 6),
				"q": "package q\nimport \"p\"\ntype B struct{ x, y p.A }\n",
			},
		},
		{
			// A valid graph, for contrast: no cycle, so no order-sensitive branch.
			name: "acyclic control",
			pkgs: map[string]string{"p": "package p\ntype T0 struct{ v int }\n" + chainOn("T0", 8)},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entry := tc.entry
			if entry == "" {
				entry = "p"
			}
			seen := map[uint64]int{}
			for range 200 {
				// Fresh fileset, resolver and checker each run, so nothing carries
				// over except the source.
				fset := token.NewFileSet()
				resolve := makeCostResolver(t, fset, tc.pkgs)
				seen[typeExpansionCost(entry, resolve(entry), resolve, nil)]++
			}
			assert.Len(t, seen, 1,
				"priced %d different ways across 200 runs (%v); the charge must not "+
					"depend on map iteration order", len(seen), seen)
			for v := range seen {
				t.Logf("%s: %d nodes", tc.name, v)
			}
		})
	}
}
