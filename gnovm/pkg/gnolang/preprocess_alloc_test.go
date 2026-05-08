package gnolang

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPreprocessAllocTestStore returns a fresh defaultStore with no
// stdlibs, plus a per-tx preprocess allocator capped at maxBytes and
// pre-wired to the supplied gas meter (mirrors keeper.AddPackage's
// setup).
func newPreprocessAllocTestStore(t *testing.T, maxBytes int64, gm stypes.GasMeter) (*defaultStore, *Allocator) {
	t.Helper()
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	preAlloc := NewAllocator(maxBytes)
	preAlloc.SetGasMeter(gm)
	st.SetPreprocessAllocator(preAlloc)
	return st, preAlloc
}

// runMemPackageRecover runs RunMemPackage and recovers any panic so the
// test can inspect the message.
func runMemPackageRecover(m *Machine, mpkg *std.MemPackage) (panicked bool, value any) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			value = r
		}
	}()
	m.RunMemPackage(mpkg, false)
	return false, nil
}

// TestPreprocessAlloc_CumulativeAcrossStatements verifies that allocations
// during preprocess accumulate across statements and panic when the
// cumulative cap is exceeded — even though no individual statement is
// near the limit.
//
// This is the core DoS protection: a 500MB cap can be drained by
// thousands of innocent-looking declarations.
func TestPreprocessAlloc_CumulativeAcrossStatements(t *testing.T) {
	st, preAlloc := newPreprocessAllocTestStore(t, 4*1024, stypes.NewInfiniteGasMeter())
	defer st.SetPreprocessAllocator(nil)

	pkgPath := "gno.land/r/test/cumulative"
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
		Alloc:   NewAllocator(64 * 1024 * 1024),
	})
	defer m.Release()

	// Many top-level decls. Each is individually tiny but the
	// cumulative allocator pressure during preprocess (type values,
	// const string allocs, block items) exceeds 4KB well before the
	// 256th decl.
	var b strings.Builder
	b.WriteString("package cumulative\n")
	for i := 0; i < 256; i++ {
		fmt.Fprintf(&b, "const C%d = \"x\"\n", i)
	}
	b.WriteString("func main() {}\n")

	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "cumulative",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
	}
	panicked, val := runMemPackageRecover(m, mpkg)
	require.True(t, panicked, "expected preprocess to exceed allocation cap")
	require.Contains(t, fmt.Sprint(val), "allocation limit exceeded",
		"panic should be alloc-limit, not something else: %v", val)

	// preAlloc should be at-or-near its cap; the outer m.Alloc should
	// have negligible usage since preprocess panicked before any of
	// the outer-machine init ran.
	maxBytes, bytes := preAlloc.Status()
	assert.Equal(t, int64(4*1024), maxBytes)
	// Bytes can be either at-cap (panic was raised exactly when the
	// next allocation overflowed) or somewhere between maxBytes-1 and
	// maxBytes. Just confirm we got close.
	assert.Greater(t, bytes, int64(2*1024),
		"preAlloc.bytes should be near the cap before panic; got %d", bytes)
}

// TestPreprocessAlloc_NoGCOnHardCap verifies the panic message
// distinguishes the no-GC hard cap from the regular GC-retry path.
// This is the protection invariant: GC during preprocess would
// undercount because GarbageCollect doesn't visit m.Values, so the
// preAlloc must NEVER attempt GC on overflow.
func TestPreprocessAlloc_NoGCOnHardCap(t *testing.T) {
	st, preAlloc := newPreprocessAllocTestStore(t, 2*1024, stypes.NewInfiniteGasMeter())
	defer st.SetPreprocessAllocator(nil)

	// Inspect: collect must remain nil after store wires up the
	// allocator. NewMachineWithOptions's isPreprocessing path skips
	// SetGCFn, so collect should still be nil here AND after
	// running preprocess (verified below).
	require.Nil(t, preAlloc.collect, "preAlloc.collect must be nil before preprocess")

	pkgPath := "gno.land/r/test/nogc"
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
		Alloc:   NewAllocator(64 * 1024 * 1024),
	})
	defer m.Release()

	var b strings.Builder
	b.WriteString("package nogc\n")
	for i := 0; i < 256; i++ {
		fmt.Fprintf(&b, "const C%d = \"x\"\n", i)
	}
	b.WriteString("func main() {}\n")

	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "nogc",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
	}
	panicked, val := runMemPackageRecover(m, mpkg)
	require.True(t, panicked, "expected preprocess to exceed allocation cap")
	// The "(no GC)" suffix is the marker that we hit the hard-cap
	// path in alloc.go, not the GC-retry-then-fail path. This
	// proves the preprocess allocator is configured collect=nil.
	require.Contains(t, fmt.Sprint(val), "(no GC)",
		"hard-cap panic must include '(no GC)' marker; got: %v", val)

	// And collect is still nil after the failed preprocess: nothing
	// in the sub-Machine setup overwrote it.
	assert.Nil(t, preAlloc.collect, "preAlloc.collect must remain nil after preprocess")
}

// TestPreprocessAlloc_InitGetsSeparateAllocator verifies the outer
// Machine's allocator is independent of the per-tx preprocess
// allocator — so a successful preprocess that uses near-cap budget
// does not starve the init() runtime budget.
//
// Concretely: install a tight preAlloc that fits a small package's
// preprocess but nothing more. After successful preprocess+init,
// preAlloc.bytes should be > 0 (preprocess did allocate) and
// m.Alloc.bytes should reflect init/runtime allocations independent
// from preAlloc.
func TestPreprocessAlloc_InitGetsSeparateAllocator(t *testing.T) {
	// Pick a comfortable preAlloc; we want preprocess to succeed.
	// 256 KB is plenty for the small fixture below.
	st, preAlloc := newPreprocessAllocTestStore(t, 256*1024, stypes.NewInfiniteGasMeter())
	defer st.SetPreprocessAllocator(nil)

	outerAlloc := NewAllocator(8 * 1024 * 1024)
	pkgPath := "gno.land/r/test/initsep"
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
		Alloc:   outerAlloc,
	})
	defer m.Release()

	// Source uses a few string consts so preprocess allocates a
	// non-zero amount via alloc.NewString during constant
	// evaluation. Without strings/types, simple int consts don't
	// hit the allocator at preprocess time.
	src := `package initsep
const A = "preprocess-time-string-A"
const B = "preprocess-time-string-B"
const C = A + B
var X = 1
var Y = X + 2
func init() { _ = X + Y }
func main() {}
`
	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "initsep",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: src}},
	}
	panicked, val := runMemPackageRecover(m, mpkg)
	require.False(t, panicked, "preprocess+init should succeed; panic: %v", val)

	_, preBytes := preAlloc.Status()
	_, outBytes := outerAlloc.Status()

	// Preprocess used some non-zero share of preAlloc.
	assert.Greater(t, preBytes, int64(0),
		"preAlloc should have non-zero bytes after preprocess")
	// Outer allocator separately tracked init/runtime allocations.
	assert.Greater(t, outBytes, int64(0),
		"outer alloc should have non-zero bytes after init+runtime")
	// The two counters are independent: outer is not inflated by
	// preAlloc usage. (We don't care which one is bigger; just that
	// they aren't aliased.)
	assert.NotEqual(t, &preAlloc.bytes, &outerAlloc.bytes,
		"preAlloc and outer alloc must be distinct objects")
}

// TestPreprocessAlloc_GasCharged verifies the preprocess allocator's
// gas meter is consumed proportional to allocation work — and that
// running out of gas mid-preprocess raises an OutOfGas error rather
// than continuing.
func TestPreprocessAlloc_GasCharged(t *testing.T) {
	// Generous bytes cap; tight gas budget. Triggers OOG via
	// gas-meter ConsumeGas, NOT via maxBytes.
	gm := stypes.NewGasMeter(50_000) // ~50k gas — small fraction of typical preprocess
	st, preAlloc := newPreprocessAllocTestStore(t, 8*1024*1024, gm)
	defer st.SetPreprocessAllocator(nil)

	pkgPath := "gno.land/r/test/gascharged"
	m := NewMachineWithOptions(MachineOptions{
		PkgPath:  pkgPath,
		Store:    st,
		Output:   io.Discard,
		Alloc:    NewAllocator(64 * 1024 * 1024),
		GasMeter: gm, // share gas meter with outer machine too
	})
	defer m.Release()

	var b strings.Builder
	b.WriteString("package gascharged\n")
	// String consts force alloc.NewString → alloc-gas charges per
	// allocation. Simple int consts don't hit the allocator.
	for i := 0; i < 1024; i++ {
		fmt.Fprintf(&b, "const C%d = \"some-non-trivial-string-%d\"\n", i, i)
	}
	b.WriteString("func main() {}\n")

	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "gascharged",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
	}
	panicked, val := runMemPackageRecover(m, mpkg)
	require.True(t, panicked, "expected OOG before preprocess completes")
	// OOG can surface either via the gas meter's OutOfGasError or a
	// generic "out of gas" string depending on the call path. Just
	// confirm gas was the proximate cause.
	msg := fmt.Sprint(val)
	require.True(t,
		strings.Contains(msg, "out of gas") || strings.Contains(msg, "OutOfGasError"),
		"expected OOG panic, got: %v", val)
	require.True(t, gm.IsPastLimit(),
		"gas meter should be past its limit; consumed=%d limit=%d",
		gm.GasConsumed(), gm.Limit())

	// preAlloc.bytes still non-zero — work was done before OOG.
	_, preBytes := preAlloc.Status()
	assert.Greater(t, preBytes, int64(0))
}

// TestPreprocessAlloc_DoublingConcatBlowsUpFromTinySource is the
// "small adversarial input, huge allocation" canonical case: each
// const string is the previous concatenated with itself, so the
// in-memory string grows 2^N from N short source lines. ~30 lines
// of source can demand >1 GB of heap during preprocess-time const
// folding. The hard-cap allocator must catch this BEFORE the
// program is loaded, no matter how generous the maxBytes.
//
// Source size: O(N). Allocated bytes: O(2^N).
func TestPreprocessAlloc_DoublingConcatBlowsUpFromTinySource(t *testing.T) {
	// 1 MB cap. Starting string is 8 bytes; we'll exceed 1 MB
	// somewhere around iteration 17 (2^17 * 8 = 1 MB). 24
	// doublings give a worst-case demand of 2^24 * 8 = 128 MB —
	// but we should panic well before generating that.
	st, _ := newPreprocessAllocTestStore(t, 1*1024*1024, stypes.NewInfiniteGasMeter())
	defer st.SetPreprocessAllocator(nil)

	pkgPath := "gno.land/r/test/doubling"
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
		Alloc:   NewAllocator(64 * 1024 * 1024),
	})
	defer m.Release()

	const N = 24
	var b strings.Builder
	b.WriteString("package doubling\n")
	b.WriteString(`const a0 = "abcdefgh"` + "\n")
	for i := 1; i <= N; i++ {
		fmt.Fprintf(&b, "const a%d = a%d + a%d\n", i, i-1, i-1)
	}
	b.WriteString("func main() {}\n")

	// Source size sanity check: should be tiny (~hundreds of bytes).
	src := b.String()
	require.Less(t, len(src), 1024,
		"source must stay small; got %d bytes", len(src))

	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "doubling",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: src}},
	}
	panicked, val := runMemPackageRecover(m, mpkg)
	require.True(t, panicked, "expected hard-cap panic from doubling concat")
	msg := fmt.Sprint(val)
	require.Contains(t, msg, "allocation limit exceeded",
		"expected alloc-cap panic, got: %v", val)
	require.Contains(t, msg, "(no GC)",
		"expected no-GC marker (preprocess hard-cap), got: %v", val)
}

// TestPreprocessAlloc_NotSetMeansSubMachineAllocNil verifies that when
// a Store has no preprocessAlloc installed, the sub-Machine fallback
// resolves to a nil Allocator (the historical default) — i.e. the
// new code does not regress non-keeper paths (filetests, REPL, etc.)
// that never call SetPreprocessAllocator.
func TestPreprocessAlloc_NotSetMeansSubMachineAllocNil(t *testing.T) {
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	require.Nil(t, st.GetPreprocessAllocator(),
		"freshly constructed store has no preprocessAlloc")

	// Sub-Machine via NewMachine(pkg, store) with no opts.Alloc and
	// no store preprocessAlloc → opts.MaxAllocBytes is 0 →
	// NewAllocator(0) returns nil.
	sub := NewMachine("test", st)
	defer sub.Release()
	assert.Nil(t, sub.Alloc,
		"sub-Machine should have nil Alloc when neither opt nor store provides one")
}

// TestPreprocessAlloc_BeginTransactionPropagates verifies the per-tx
// preprocess allocator survives BeginTransaction. Sub-Machines created
// inside Preprocess fork the store via BeginTransaction first
// (preprocess.go:3944) — without propagation, they would see a fresh
// store with preprocessAlloc = nil and no allocations would be
// counted.
func TestPreprocessAlloc_BeginTransactionPropagates(t *testing.T) {
	st, preAlloc := newPreprocessAllocTestStore(t, 1024*1024, stypes.NewInfiniteGasMeter())
	defer st.SetPreprocessAllocator(nil)

	tx := st.BeginTransaction(nil, nil, nil, nil)
	got := tx.GetPreprocessAllocator()
	require.NotNil(t, got, "forked tx-store must inherit preprocessAlloc")
	require.Same(t, preAlloc, got,
		"forked tx-store must share the SAME *Allocator pointer (gas counters and bytes are shared across the tx)")
}
