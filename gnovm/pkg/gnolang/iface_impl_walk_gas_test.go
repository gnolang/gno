package gnolang

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// Gas metering of the interface-satisfaction BFS (InterfaceType.checkImplementedBy
// → embedWalk), at preprocess and at runtime.
//
// Preprocess: every assignability check against a non-empty interface walks the
// concrete type's embedding graph; before the fix only the flat per-byte
// preprocess gas applied. Runtime: a type assertion x.(I), a two-value
// assertion x, ok := e.(I), a type switch `case I:`, and lazy interface method
// values run the same walk; before the fix only the per-method part was
// charged. These tests pin that the walk is billed against the tx gas meter on
// every such path.
//
// Which test pins which charge: the wide-interface fixtures
// (TypeAssert/TypeSwitch/ScalesWithIterations) are dominated by the per-scan
// charge (OpCPUSlopeEmbedScan) and fail if it is removed; the 1-method
// fixtures (the Preprocess_* tests, LazyBound, MethodValueBind) are dominated
// by the per-type expansion charge (OpCPUSlopeEmbedExpand) and fail if
// that is removed; TestRuntime_TrailHop_Charged pins the per-hop trail charge
// (OpCPUSlopeEmbedTrailHop) on a deep chain.

// buildWideEmbedPkg emits a package with a 1-method interface I and a struct S
// embedding nEmbed field-less types of which only the last provides the method
// — so S satisfies I only through its embedding graph and every satisfaction
// check (or method lookup) walks all nEmbed embedded types — followed by tail,
// the statements exercising a particular path. The single method keeps the
// per-method charge negligible so tests discriminate the per-field charge.
func buildWideEmbedPkg(pkgName string, nEmbed int, tail string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	b.WriteString("type I interface{ M() }\n\n")
	for i := range nEmbed {
		fmt.Fprintf(&b, "type T%d struct{}\n", i)
	}
	fmt.Fprintf(&b, "func (T%d) M() {}\n", nEmbed-1) // unique provider → walk expands all
	b.WriteString("\ntype S struct {\n")
	for i := range nEmbed {
		fmt.Fprintf(&b, "\tT%d\n", i)
	}
	b.WriteString("}\n\n")
	b.WriteString(tail)
	return b.String()
}

// buildRuntimeAssertPkg returns a package whose init() performs nAssert
// interface checks over a value of a struct S that embeds nEmbed distinct types
// (each providing one method of the nEmbed-method interface I). The assignment
// `var e any = S{}` is an empty-interface conversion, so preprocess stays cheap
// and the metered work happens at runtime inside init().
func buildRuntimeAssertPkg(pkgName string, nEmbed, nAssert int, useSwitch bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	b.WriteString("type I interface {\n")
	for i := range nEmbed {
		fmt.Fprintf(&b, "\tM%d()\n", i)
	}
	b.WriteString("}\n\n")

	for i := range nEmbed {
		fmt.Fprintf(&b, "type T%d struct{}\n", i)
		fmt.Fprintf(&b, "func (T%d) M%d() {}\n", i, i)
	}
	b.WriteString("\n")

	b.WriteString("type S struct {\n")
	for i := range nEmbed {
		fmt.Fprintf(&b, "\tT%d\n", i)
	}
	b.WriteString("}\n\n")

	b.WriteString("var e any = S{}\n")
	b.WriteString("var sink bool\n\n")

	b.WriteString("func init() {\n")
	fmt.Fprintf(&b, "\tfor i := 0; i < %d; i++ {\n", nAssert)
	if useSwitch {
		b.WriteString("\t\tswitch e.(type) {\n\t\tcase I:\n\t\t\tsink = true\n\t\tdefault:\n\t\t\tsink = false\n\t\t}\n")
	} else {
		b.WriteString("\t\t_, ok := e.(I)\n\t\tsink = ok\n")
	}
	b.WriteString("\t}\n}\n\n")

	b.WriteString("func main() {}\n")
	return b.String()
}

// runPkgSource deploys and runs a single-file package under a gas meter of the
// given budget, with the keeper's preprocess-allocator wiring (so preprocess
// and runtime bill the same meter). Returns whether it panicked, the panic
// value, and the gas consumed.
func runPkgSource(t *testing.T, pkgName string, budget int64, src string) (panicked bool, val any, gasUsed int64) {
	t.Helper()
	gm := stypes.NewGasMeter(budget)
	st, _ := newPreprocessAllocTestStore(t, 512*1024*1024, gm)
	defer st.SetPreprocessAllocator(nil)

	pkgPath := "gno.land/r/test/" + pkgName
	m := NewMachineWithOptions(MachineOptions{
		PkgPath:  pkgPath,
		Store:    st,
		Output:   io.Discard,
		Alloc:    NewAllocator(512 * 1024 * 1024),
		GasMeter: gm,
	})
	defer m.Release()
	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  pkgName,
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: src}},
	}
	panicked, val = runMemPackageRecover(m, mpkg)
	return panicked, val, gm.GasConsumed()
}

// runRuntimeAssertPkg runs buildRuntimeAssertPkg's package under budget.
func runRuntimeAssertPkg(t *testing.T, pkgName string, budget int64, nEmbed, nAssert int, useSwitch bool) (bool, any, int64) {
	t.Helper()
	return runPkgSource(t, pkgName, budget, buildRuntimeAssertPkg(pkgName, nEmbed, nAssert, useSwitch))
}

// requireOOG asserts that a run panicked with an out-of-gas error.
func requireOOG(t *testing.T, panicked bool, val any) {
	t.Helper()
	require.True(t, panicked, "expected OOG from the metered interface-satisfaction walk; "+
		"if this passes, the per-field BFS work is unmetered")
	msg := fmt.Sprint(val)
	require.True(t,
		strings.Contains(msg, "out of gas") || strings.Contains(msg, "OutOfGasError"),
		"expected OOG panic, got: %v", val)
}

// ---- preprocess ----
//
// Budget arithmetic (shared by the three *_Metered tests): the fixture from
// buildWideEmbedPkg has a 1-method interface and a struct embedding 64 types
// of which only the last provides the method, so every check expands and
// scans all 64. Per check that is OpCPUSlopeTypeAssertIface × 1 (349 gas,
// negligible) plus 64 × (OpCPUSlopeEmbedExpand + OpCPUSlopeEmbedScan) + one
// OpCPUSlopeEmbedTrailHop ≈ 14.9K gas. 20 checks charge ≈300K gas for the walk
// on top of ≈340K allocation gas, so a 500K budget OOGs. Without the per-type
// expansion charge the package costs ≈340K + 20 × (64 × 25 + 135 + 349) ≈ 382K
// and fits, so these tests fail if that charge alone is removed (verified by
// mutation) — a wide-interface fixture would not: 64 methods × 20 checks × 349
// already exceeds 500K by itself.

// TestPreprocess_IfaceImpl_GasCharged: the assignment form
// `var x I = S{}` (checkOrConvertType → mustAssignableTo).
func TestPreprocess_IfaceImpl_GasCharged(t *testing.T) {
	var tail strings.Builder
	for i := range 20 {
		fmt.Fprintf(&tail, "var x%d I = S{}\n", i)
	}
	tail.WriteString("func main() {}\n")
	panicked, val, _ := runPkgSource(t, "verifyimpl", 500_000, buildWideEmbedPkg("verifyimpl", 64, tail.String()))
	requireOOG(t, panicked, val)
}

// TestPreprocess_IfaceImpl_GasSufficient verifies that with a
// generous gas budget, the package preprocesses successfully — confirming
// the gas charge is not so high as to reject legitimate code.
func TestPreprocess_IfaceImpl_GasSufficient(t *testing.T) {
	body := buildWideEmbedPkg("verifyimpl2", 16, "var x I = S{}\nfunc main() {}\n")
	panicked, val, _ := runPkgSource(t, "verifyimpl2", 50_000_000, body)
	require.False(t, panicked, "expected preprocess to succeed with generous gas budget, got panic: %v", val)
}

// TestPreprocess_IfaceImpl_ConversionMetered: the explicit
// conversion form I(S{}), which resolves satisfaction through the CallExpr
// interface-conversion branch and convertConst rather than checkOrConvertType.
// A conversion runs the check twice (both branches), so 20 statements are 40
// walks: ≈600K gas of walk charge, or ≈83K without the expansion charge, on
// top of ≈340K allocation gas. The 600K budget therefore OOGs only with the
// per-type expansion charge present.
func TestPreprocess_IfaceImpl_ConversionMetered(t *testing.T) {
	var tail strings.Builder
	for i := range 20 {
		fmt.Fprintf(&tail, "var x%d = I(S{})\n", i)
	}
	tail.WriteString("func main() {}\n")
	panicked, val, _ := runPkgSource(t, "cvtimpl", 600_000, buildWideEmbedPkg("cvtimpl", 64, tail.String()))
	requireOOG(t, panicked, val)
}

// TestPreprocess_IfaceImpl_MultiAssignMetered: the multi-value
// assignment form `a, b = f()` with interface destinations, which resolves
// satisfaction through AssignStmt.AssertCompatible. 20 statements × 2
// destinations = 40 walks ≈ 600K gas of walk charge against 500K.
func TestPreprocess_IfaceImpl_MultiAssignMetered(t *testing.T) {
	var tail strings.Builder
	tail.WriteString("func f() (S, S) { return S{}, S{} }\n")
	tail.WriteString("var a, b I\n")
	tail.WriteString("func main() {\n")
	for range 20 {
		tail.WriteString("\ta, b = f()\n")
	}
	tail.WriteString("}\n")
	panicked, val, _ := runPkgSource(t, "multiimpl", 500_000, buildWideEmbedPkg("multiimpl", 64, tail.String()))
	requireOOG(t, panicked, val)
}

// ---- runtime ----

// TestRuntime_TypeAssert_GasCharged: a tight budget that comfortably covers the
// one-time preprocess+deploy but not the runtime assertion loop must OOG,
// proving the runtime assertion walk is metered. Per assertion over the 48×48
// shape: 48 × 349 + 48 × OpCPUSlopeEmbedExpand + 48² × OpCPUSlopeEmbedScan +
// 48 × OpCPUSlopeEmbedTrailHop ≈ 90K gas, so 100 assertions ≈ 9M against a
// 5M budget.
func TestRuntime_TypeAssert_GasCharged(t *testing.T) {
	panicked, val, _ := runRuntimeAssertPkg(t, "rtassert", 5_000_000, 48, 100, false)
	requireOOG(t, panicked, val)
}

// TestRuntime_TypeSwitch_GasCharged: the same for a `case I:` type switch, which
// runs the identical walk.
func TestRuntime_TypeSwitch_GasCharged(t *testing.T) {
	panicked, val, _ := runRuntimeAssertPkg(t, "rtswitch", 5_000_000, 48, 100, true)
	requireOOG(t, panicked, val)
}

// TestRuntime_TypeAssert_GasSufficient: with a generous budget the same package
// runs to completion — the charge is not so high as to reject legitimate code.
func TestRuntime_TypeAssert_GasSufficient(t *testing.T) {
	panicked, val, _ := runRuntimeAssertPkg(t, "rtok", 2_000_000_000, 48, 8, false)
	require.False(t, panicked, "expected success with a generous budget, got panic: %v", val)
}

// TestRuntime_TypeAssert_ScalesWithIterations proves the charge is per-runtime-
// assertion (not a one-time preprocess cost): under an ample budget, running the
// loop 8 times consumes strictly and substantially more gas than running it
// once. The one-time preprocess/deploy cost cancels in the delta, so the extra
// gas is exactly the 7 additional runtime assertions' metered BFS work.
func TestRuntime_TypeAssert_ScalesWithIterations(t *testing.T) {
	const nEmbed = 48
	p1, _, g1 := runRuntimeAssertPkg(t, "rtscale1", 2_000_000_000, nEmbed, 1, false)
	p8, _, g8 := runRuntimeAssertPkg(t, "rtscale8", 2_000_000_000, nEmbed, 8, false)
	require.False(t, p1)
	require.False(t, p8)

	// Each assertion expands nEmbed types once, scans them for each of nEmbed
	// methods, and rebuilds a 1-hop trail per method, so per-assertion runtime
	// cost ≈ nEmbed × Expand + nEmbed² × Scan + nEmbed × TrailHop. The
	// 7-assertion delta must dwarf any noise; use a conservative lower bound
	// of half that.
	delta := g8 - g1
	perAssert := int64(nEmbed)*int64(OpCPUSlopeEmbedExpand) + int64(nEmbed)*int64(nEmbed)*int64(OpCPUSlopeEmbedScan) + int64(nEmbed)*int64(OpCPUSlopeEmbedTrailHop)
	lower := 7 * perAssert / 2
	require.Greater(t, delta, lower,
		"runtime gas did not scale with assertion count (g1=%d g8=%d delta=%d lower=%d); "+
			"per-assertion BFS walk may be unmetered", g1, g8, delta, lower)
}

// TestRuntime_TypeAssert_PerMethodChargedOnce pins the split in
// doOpTypeAssert2: a non-concrete operand (its dynamic type is itself an
// interface) fails fast and pays OpCPUSlopeTypeAssertIface per method via
// incrCPU, while a concrete operand pays the same per-method amount inside
// checkImplementedBy — and, for a type providing the methods directly (depth-0
// hits, no walk), nothing more. Either path charging twice would show here.
func TestRuntime_TypeAssert_PerMethodChargedOnce(t *testing.T) {
	const nMethods = 10
	gm := stypes.NewGasMeter(1_000_000_000)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "test", Output: io.Discard, Alloc: NewAllocator(1 << 20), GasMeter: gm,
	})
	defer m.Release()
	iface, dt, sv := benchInterfaceAndImpl(m.Alloc, nMethods, nMethods)
	other := &InterfaceType{PkgPath: "test", Methods: []FieldType{{Name: "Other", Type: &FuncType{}}}}
	perMethod := int64(OpCPUSlopeTypeAssertIface) * nMethods

	assert2 := func(x TypedValue) (ok bool, gas int64) {
		before := gm.GasConsumed()
		m.PushValue(x)
		m.PushValue(asValue(iface))
		m.PushExpr(&TypeAssertExpr{HasOK: true})
		m.doOpTypeAssert2()
		ok = m.PeekValue(1).GetBool()
		m.Values = m.Values[:0]
		return ok, gm.GasConsumed() - before
	}

	ok, gas := assert2(TypedValue{T: other}) // non-concrete: dynamic type is an interface
	require.False(t, ok)
	require.Equal(t, perMethod, gas, "non-concrete path: per-method charge only")

	ok, gas = assert2(TypedValue{T: dt, V: sv}) // concrete, methods found at depth 0
	require.True(t, ok)
	require.Equal(t, perMethod, gas, "concrete path: per-method charge once, no walk")
}

// TestErrorCheck_Metered pins that the error-interface check used in result
// formatting (IsErrorType / ImplError, reached from
// gno.land/pkg/sdk/vm/convert.go) meters its embedding-graph walk. A wide
// struct with no Error() method forces the walk over all embedded types.
func TestErrorCheck_Metered(t *testing.T) {
	const N = 128
	dt := wideEmbedDeclaredType(N)
	tight := int64(OpCPUSlopeEmbedExpand) * int64(N) / 2

	require.Panics(t, func() { IsErrorType(stypes.NewGasMeter(tight), dt) },
		"IsErrorType should OOG on a wide-embedding type under a tight budget")
	require.Panics(t, func() { (&TypedValue{T: dt}).ImplError(stypes.NewGasMeter(tight)) },
		"ImplError should OOG on a wide-embedding type under a tight budget")

	// Generous budget completes and reports not-an-error; nil meter is a no-op.
	require.False(t, IsErrorType(stypes.NewGasMeter(1_000_000_000), dt))
	require.False(t, (&TypedValue{T: dt}).ImplError(stypes.NewGasMeter(1_000_000_000)))
	require.False(t, IsErrorType(nil, dt))
}

// TestRuntime_TrailHop_Charged pins the per-hop trail charge: a method found
// through an 8-deep embedding chain (MaxEmbedDepth) costs 8 × TrailHop more
// than the same lookup at depth 1 with the same number of types expanded and
// scanned. Uses checkImplementedBy directly so the delta is exact.
func TestRuntime_TrailHop_Charged(t *testing.T) {
	iface := &InterfaceType{PkgPath: "t", Methods: []FieldType{{Name: "M", Type: &FuncType{Params: []FieldType{}, Results: []FieldType{}}}}}
	// chain(d): S embeds L1, L1 embeds L2, …, L(d) declares M; the other d-1
	// slots of S are padded with method-less types so every shape expands
	// and scans exactly 8 types.
	chain := func(d int) *DeclaredType {
		var leaf *DeclaredType
		var build func(level int) *DeclaredType
		build = func(level int) *DeclaredType {
			st := &StructType{PkgPath: "t", Fields: []FieldType{}}
			dt := &DeclaredType{PkgPath: "t", Name: Name(fmt.Sprintf("L%d", level)), Base: st}
			if level < d {
				child := build(level + 1)
				st.Fields = []FieldType{{Name: child.Name, Type: child, Embedded: true}}
			} else {
				leaf = dt
			}
			return dt
		}
		fields := []FieldType{{Name: "L1", Type: build(1), Embedded: true}}
		for i := d; i < 8; i++ {
			pad := &DeclaredType{PkgPath: "t", Name: Name(fmt.Sprintf("P%d", i)), Base: &StructType{PkgPath: "t", Fields: []FieldType{}}}
			fields = append(fields, FieldType{Name: pad.Name, Type: pad, Embedded: true})
		}
		ft := &FuncType{Params: []FieldType{{Name: "self", Type: leaf}}, Results: []FieldType{}}
		leaf.Methods = []TypedValue{{T: ft, V: &FuncValue{Type: ft, IsMethod: true, Source: &FuncDecl{}, Name: "M", PkgPath: "t", body: []Stmt{}}}}
		return &DeclaredType{PkgPath: "t", Name: "S", Base: &StructType{PkgPath: "t", Fields: fields}}
	}
	gas := func(d int) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		require.NoError(t, iface.checkImplementedBy(gm, chain(d)))
		return gm.GasConsumed()
	}
	// Both shapes expand 8 types; depth 8 scans 8 entries across 8 levels
	// while depth 1 scans 8 entries in one level, so scan gas is equal and
	// the delta is exactly the 7 extra trail hops.
	require.Equal(t, int64(7*OpCPUSlopeEmbedTrailHop), gas(8)-gas(1))
}

// TestSatisfactionCheck_PrimitiveRoot_NoWalkState pins the walk's second
// allocation-free exit: a root that exposes no struct to expand (rootSt ==
// nil) returns before `seen` and `levels` are built. Fprint probes Stringer
// and then error on every printed value, so this is the hottest satisfaction
// path in the chain, and it carries no gas signal — dropping the early return
// costs 4 allocations and 344 B per probe while charging exactly the same, so
// only an allocation assertion can catch it.
func TestSatisfactionCheck_PrimitiveRoot_NoWalkState(t *testing.T) {
	got := testing.AllocsPerRun(1000, func() {
		// Exactly what Fprint does per printed value.
		_ = isImplementedBy(nil, gStringerType, IntType)
		_ = isImplementedBy(nil, gErrorType, IntType)
	})
	// 6 with the early return, 14 without.
	require.LessOrEqual(t, got, float64(8),
		"a primitive root must not allocate the embedWalk seen map and level slices")
}

// TestCheckImplementedBy_WalkSharedAcrossMethods pins the headline property of
// the rewrite: one embedWalk serves every method lookup of a check, so the
// embedding graph is expanded once per check instead of once per method. The
// expansion term (nEmbed × OpCPUSlopeEmbedExpand) must appear exactly once in
// the total. If newEmbedWalk moved back inside checkImplementedBy's method
// loop — the pre-PR per-method re-expansion that these constants were re-sized
// for — that term would be multiplied by the method count and this assertion
// fails. Every other walk-gas assertion in this file is one-sided (requireOOG,
// or require.Greater against a half-magnitude floor) and the two exact-equality
// tests use shapes where sharing is vacuous (depth-0 methods, or one method),
// so without this test that regression is caught by nothing.
func TestCheckImplementedBy_WalkSharedAcrossMethods(t *testing.T) {
	const (
		nMethods = 16
		nEmbed   = 8
	)
	methods := make([]FieldType, nMethods)
	for i := range nMethods {
		methods[i] = FieldType{
			Name: Name(fmt.Sprintf("M%d", i)),
			Type: &FuncType{Params: []FieldType{}, Results: []FieldType{}},
		}
	}
	iface := &InterfaceType{PkgPath: "t", Methods: methods}

	// S embeds nEmbed method-less types except the last, which provides every
	// method: each lookup misses at depth 0, scans all nEmbed entries of the
	// single level, and rebuilds a 1-hop trail.
	fields := make([]FieldType, nEmbed)
	for i := range nEmbed {
		dt := &DeclaredType{
			PkgPath: "t", Name: Name(fmt.Sprintf("T%d", i)),
			Base: &StructType{PkgPath: "t", Fields: []FieldType{}},
		}
		fields[i] = FieldType{Name: dt.Name, Type: dt, Embedded: true}
	}
	provider := fields[nEmbed-1].Type.(*DeclaredType)
	for i := range nMethods {
		ft := &FuncType{Params: []FieldType{{Name: "self", Type: provider}}, Results: []FieldType{}}
		provider.Methods = append(provider.Methods, TypedValue{T: ft, V: &FuncValue{
			Type: ft, IsMethod: true, Source: &FuncDecl{},
			Name: methods[i].Name, PkgPath: "t", body: []Stmt{},
		}})
	}
	root := &DeclaredType{PkgPath: "t", Name: "S", Base: &StructType{PkgPath: "t", Fields: fields}}

	gm := stypes.NewGasMeter(1_000_000_000)
	require.NoError(t, iface.checkImplementedBy(gm, root))

	want := int64(nMethods)*int64(OpCPUSlopeTypeAssertIface) + // per method
		int64(nEmbed)*int64(OpCPUSlopeEmbedExpand) + // expansion: ONCE per check
		int64(nMethods)*int64(nEmbed)*int64(OpCPUSlopeEmbedScan) + // one level scan per method
		int64(nMethods)*int64(OpCPUSlopeEmbedTrailHop) // one 1-hop trail per hit
	require.Equal(t, want, gm.GasConsumed(),
		"the embedding graph must be expanded once per check, not once per method lookup")
}

// lazyBoundTail takes a method value off an interface value (`var fn = i.M`, a
// lazy interface bind with Func==nil) and calls it nCalls times. Each call
// resolves the receiver at call time via resolveLazyBound, which walks the
// embedding graph of S through findEmbeddedFieldType.
func lazyBoundTail(nCalls int) string {
	return fmt.Sprintf("var i I = S{}\nvar fn = i.M\nvar sink int\n\n"+
		"func init() {\n\tfor k := 0; k < %d; k++ {\n\t\tfn()\n\t\tsink++\n\t}\n}\n\nfunc main() {}\n", nCalls)
}

// TestRuntime_LazyBoundResolve_GasCharged: calling a lazy interface method value
// over a wide-embedding receiver walks the embedding graph per call; a tight
// budget that covers the cheap deploy but not the call loop must OOG. Without the
// per-field metering this walk is free (only OpCPULazyBoundResolve per hop), so
// the loop would fit the budget — this test fails unless the walk is billed.
func TestRuntime_LazyBoundResolve_GasCharged(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "lazybound", 5_000_000, buildWideEmbedPkg("lazybound", 64, lazyBoundTail(500)))
	requireOOG(t, panicked, val)
}

// TestRuntime_LazyBoundResolve_GasSufficient: a generous budget completes.
func TestRuntime_LazyBoundResolve_GasSufficient(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "lazyok", 5_000_000_000, buildWideEmbedPkg("lazyok", 64, lazyBoundTail(500)))
	require.False(t, panicked, "expected success with a generous budget, got panic: %v", val)
}

// methodValueBindTail forms a method value off an interface value (i.M) inside
// a loop. Each formation runs getPointerToFromTV's VPInterface branch, which
// walks the concrete receiver's embedding graph via findEmbeddedFieldType
// (bind-time, distinct from the call-time resolveLazyBound walk).
func methodValueBindTail(nForms int) string {
	return fmt.Sprintf("var i I = S{}\nvar sink func()\n\n"+
		"func init() {\n\tfor k := 0; k < %d; k++ {\n\t\tf := i.M\n\t\tsink = f\n\t}\n}\n\nfunc main() {}\n", nForms)
}

// TestRuntime_MethodValueBind_GasCharged: forming a lazy interface method value
// over a wide-embedding receiver walks the embedding graph at bind time; a tight
// budget must OOG. Fails unless getPointerToFromTV's walk is billed.
func TestRuntime_MethodValueBind_GasCharged(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "mvbind", 5_000_000, buildWideEmbedPkg("mvbind", 64, methodValueBindTail(500)))
	requireOOG(t, panicked, val)
}

// TestRuntime_MethodValueBind_GasSufficient: a generous budget completes.
func TestRuntime_MethodValueBind_GasSufficient(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "mvbindok", 5_000_000_000, buildWideEmbedPkg("mvbindok", 64, methodValueBindTail(500)))
	require.False(t, panicked, "expected success with a generous budget, got panic: %v", val)
}
