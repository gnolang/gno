package gnolang

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gasprof"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// net returns the reconcilable gas total (gross dimensions minus refunds).
func netGas(tot gasprof.Totals) int64 {
	return tot.CPU + tot.Alloc + tot.Store + tot.Other - tot.Refund
}

func newGasProfMachine(t *testing.T, pkgPath, src string) *Machine {
	t.Helper()
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	gnoStore := NewStore(nil, baseStore, iavlStore)
	m := NewMachine(pkgPath, gnoStore)
	// The profiler observes a GasMeter, so wire one (an infinite meter never
	// affects execution). This is the CPU + alloc seam the dev surface uses.
	m.GasMeter = stypes.NewInfiniteGasMeter()
	m.RunMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  pkgPath[strings.LastIndexByte(pkgPath, '/')+1:],
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "gas.gno", Body: src}},
	}, true)
	return m
}

// Attribution test: run a program with a KNOWN call structure (method,
// recursion, closure, plain func) and assert the profiler produces the right
// qualified frame names, captures both CPU and alloc gas, and reconciles
// exactly with the gas meter. Keep the workload small (fib(8)) so it's fast.
const gasProfAttrSrc = `package attr

type Tree struct{ n int }

func (t *Tree) Insert(v int) { t.n += triangular(v) }

func triangular(v int) int {
	s := 0
	for i := 0; i < v; i++ {
		s += i
	}
	return s
}

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func Run() []int {
	t := &Tree{}
	t.Insert(30)
	square := func(x int) int { return x * x }
	out := make([]int, 0, 8) // force some allocation gas
	out = append(out, t.n, fib(8), square(9))
	return out
}
`

func TestGasProf_attributionAndConservation(t *testing.T) {
	const pkgPath = "gno.land/r/demo/attr"
	m := newGasProfMachine(t, pkgPath, gasProfAttrSrc)
	defer m.Release()

	gc0 := m.GasMeter.GasConsumed()
	prof := m.EnableGasProfiler()
	require.NotNil(t, prof, "machine must have a meter to profile")
	res := m.Eval(Call(X("Run")))
	m.DisableGasProfiler()
	gc1 := m.GasMeter.GasConsumed()

	require.Len(t, res, 1)

	var fb strings.Builder
	require.NoError(t, prof.WriteFolded(&fb))
	folded := fb.String()

	// Qualified frame identities the machine must produce.
	require.Contains(t, folded, "gno.land/r/demo/attr.Run")
	require.Contains(t, folded, "gno.land/r/demo/attr.triangular")
	require.Contains(t, folded, "gno.land/r/demo/attr.fib")
	require.Contains(t, folded, "(*gno.land/r/demo/attr.Tree).Insert") // pointer method
	require.Contains(t, folded, "gno.land/r/demo/attr.(anonymous)")    // closure
	// Full call chain Run -> Insert -> triangular attributed correctly.
	require.Contains(t, folded,
		"gno.land/r/demo/attr.Run;(*gno.land/r/demo/attr.Tree).Insert;gno.land/r/demo/attr.triangular")
	// Recursion shows as repeated fib frames within a single stack.
	require.Regexp(t, `attr\.fib;gno\.land/r/demo/attr\.fib`, folded)

	// Both dimensions captured in Phase 1; store is not wired on this surface.
	tot := prof.Totals()
	require.Positive(t, tot.CPU, "cpu gas captured")
	require.Positive(t, tot.Alloc, "alloc gas captured")
	require.Zero(t, tot.Store, "store gas is not charged on the test surface (Phase 2)")

	// Reconciliation invariant: the profile's dimensions sum exactly to the gas
	// the meter recorded for this call. Nothing dropped, nothing double-counted.
	require.Equal(t, gc1-gc0, netGas(tot), "profile must reconcile with the meter")
}

// Defer + recovered panic drives the O(1) Pop path through unwinding: the
// deferred closure and the function it calls must be attributed to the defer
// frame (not stranded), and the cursor must survive the recovery so the
// profile still reconciles. Reconciliation alone is cursor-blind, so we also
// assert the attribution chains.
const gasProfDeferSrc = `package deferp

func leaf() int {
	s := 0
	for i := 0; i < 60; i++ {
		s += i
	}
	panic("boom")
}

func cleanup() int {
	s := 0
	for i := 0; i < 40; i++ {
		s += i
	}
	return s
}

func guarded() (out int) {
	defer func() {
		if r := recover(); r != nil {
			out = cleanup()
		}
	}()
	return leaf()
}

func Run() int { return guarded() }
`

func TestGasProf_deferAndRecoveredPanic(t *testing.T) {
	m := newGasProfMachine(t, "gno.land/r/demo/deferp", gasProfDeferSrc)
	defer m.Release()

	gc0 := m.GasMeter.GasConsumed()
	prof := m.EnableGasProfiler()
	require.NotNil(t, prof)
	res := m.Eval(Call(X("Run")))
	m.DisableGasProfiler()
	gc1 := m.GasMeter.GasConsumed()
	require.Len(t, res, 1)

	var fb strings.Builder
	require.NoError(t, prof.WriteFolded(&fb))
	folded := fb.String()

	// leaf's pre-panic charge landed on leaf, under guarded.
	require.Contains(t, folded, "deferp.guarded;gno.land/r/demo/deferp.leaf")
	// The deferred closure ran and charged gas under guarded (not stranded on leaf).
	require.Contains(t, folded, "deferp.guarded;gno.land/r/demo/deferp.(anonymous)")
	// cleanup, called from the deferred closure, is attributed to it.
	require.Contains(t, folded, "deferp.(anonymous);gno.land/r/demo/deferp.cleanup")
	// Cursor survived the unwind: the profile still reconciles.
	require.Equal(t, gc1-gc0, netGas(prof.Totals()))
}

// A finite meter that runs out mid-execution must still reconcile — the final
// aborting charge is recorded before the meter panics.
func TestGasProf_outOfGasReconciles(t *testing.T) {
	const pkgPath = "gno.land/r/demo/attr"

	// Measure the full cost with an infinite meter.
	m1 := newGasProfMachine(t, pkgPath, gasProfAttrSrc)
	p1 := m1.EnableGasProfiler()
	require.NotNil(t, p1)
	m1.Eval(Call(X("Run")))
	total := m1.GasMeter.GasConsumed()
	m1.Release()
	require.Positive(t, total)

	// Re-run with a finite meter set to run out partway through.
	m := newGasProfMachine(t, pkgPath, gasProfAttrSrc)
	defer m.Release()
	m.GasMeter = stypes.NewGasMeter(total / 2)
	gc0 := m.GasMeter.GasConsumed()
	prof := m.EnableGasProfiler()
	require.NotNil(t, prof)
	func() {
		defer func() { _ = recover() }() // swallow OutOfGasError
		m.Eval(Call(X("Run")))
	}()
	// Disable didn't run (panic), so read through the still-wrapped meter.
	gc1 := m.GasMeter.GasConsumed()

	require.Greater(t, gc1, gc0, "meter consumed gas")
	require.LessOrEqual(t, gc1, total, "did not exceed the full cost")
	require.Equal(t, gc1-gc0, netGas(prof.Totals()),
		"out-of-gas profile must include the final charge and reconcile")
}

// Alloc gas must land on the allocating function, and CPU gas on the crunching
// one — reconciliation can't catch mis-attribution between dimensions/functions.
const gasProfAllocSrc = `package allocp

func allocator() int {
	a := make([]int, 500)
	return len(a)
}

func cruncher() int {
	s := 0
	for i := 0; i < 500; i++ {
		s += i
	}
	return s
}

func Run() int { return allocator() + cruncher() }
`

func TestGasProf_dimensionAttribution(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	m := newGasProfMachine(t, "gno.land/r/demo/allocp", gasProfAllocSrc)
	defer m.Release()
	prof := m.EnableGasProfiler()
	require.NotNil(t, prof)
	m.Eval(Call(X("Run")))
	m.DisableGasProfiler()
	require.Positive(t, prof.Totals().Alloc)

	dir := t.TempDir()
	path := filepath.Join(dir, "gas.pprof")
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, prof.WritePprof(f))
	require.NoError(t, f.Close())

	top := func(index string) string {
		out, err := exec.Command(goBin, "tool", "pprof", "-top", "-flat",
			"-sample_index="+index, "-nodecount=40", path).CombinedOutput()
		require.NoError(t, err, "%s", out)
		return string(out)
	}
	// allocator dominates the alloc dimension; cruncher dominates cpu.
	require.Contains(t, top("alloc_gas"), "allocp.allocator")
	require.Contains(t, top("cpu_gas"), "allocp.cruncher")
}

// DisableGasProfiler must restore both the machine meter and the allocator's
// original meter, leaving accounting identical to a never-profiled machine.
func TestGasProf_disableRestoresMeters(t *testing.T) {
	m := newGasProfMachine(t, "gno.land/r/demo/attr", gasProfAttrSrc)
	defer m.Release()

	origGasMeter := m.GasMeter
	origAllocMeter := m.Alloc.GetGasMeter() // nil on the test surface

	m.EnableGasProfiler()
	require.NotSame(t, origGasMeter, m.GasMeter, "meter is wrapped while profiling")
	m.Eval(Call(X("Run")))
	m.DisableGasProfiler()

	require.Same(t, origGasMeter, m.GasMeter, "machine meter restored exactly")
	require.Equal(t, origAllocMeter, m.Alloc.GetGasMeter(), "alloc meter restored exactly")
	_, stillWrapped := m.GasMeter.(interface{ Unwrap() stypes.GasMeter })
	require.False(t, stillWrapped, "wrapper removed")
}
