package gnolang

import (
	"fmt"
	"io"
	"math"
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
	for i := range 256 {
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
// This is the protection invariant: preAlloc is shared by every
// preprocess sub-Machine in the tx, so a collect bound to one machine
// would walk only that machine's roots and free the rest of the tx's
// preprocess allocations on paper. The preAlloc must therefore NEVER
// attempt GC on overflow.
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
	for i := range 256 {
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
	for i := range 1024 {
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

// TestPreprocessAlloc_NotSetMeansSubMachineHasFallbackAlloc verifies
// that when a Store has no preprocessAlloc installed AND no Alloc
// option is supplied, the sub-Machine falls back to a non-nil
// MaxInt64-budget Allocator (interrealm v2: Machine.Alloc must be
// non-nil so PkgID stamping always fires).
func TestPreprocessAlloc_NotSetMeansSubMachineHasFallbackAlloc(t *testing.T) {
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	require.Nil(t, st.GetPreprocessAllocator(),
		"freshly constructed store has no preprocessAlloc")

	// Sub-Machine via NewMachine(pkg, store) with no opts.Alloc and
	// no store preprocessAlloc → falls back to NewAllocator(math.MaxInt64).
	sub := NewMachine("test", st)
	defer sub.Release()
	require.NotNil(t, sub.Alloc,
		"sub-Machine Alloc must be non-nil (interrealm v2: mandatory for stamping)")
	maxBytes, _ := sub.Alloc.Status()
	require.Equal(t, int64(math.MaxInt64), maxBytes,
		"fallback Alloc should have MaxInt64 budget (no enforcement)")
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

// TestCheckArrayAllocFits exercises the preprocess-time array-length guard
// directly. The make20/21/22 filetests only cover the wildly-oversized
// MaxInt64 case end-to-end; the cases that matter for the threshold math —
// the exact boundary, and the per-element divergence between the byte
// (1 byte/elem) and non-byte (allocArrayItem/elem) paths — can't be written
// as filetests because a boundary-length array would attempt a real
// allocation. They are checked here instead.
func TestCheckArrayAllocFits(t *testing.T) {
	// Thresholds derived the same way the guard does, so the boundary cases
	// double as change-detectors for the formula.
	perItem := int64(allocArrayItem) // non-byte: a full TypedValue slot.
	thrItem := (math.MaxInt64 - allocArray) / perItem
	thrByte := int64(math.MaxInt64-allocArray) / 1 // byte: 1 byte/elem.

	tests := []struct {
		name    string
		et      Type
		length  int64
		wantMsg string // "" => must not panic
	}{
		{"zero length", IntType, 0, ""},
		{"negative length", IntType, -1, ""},
		{"small array", IntType, 1 << 20, ""},
		{"non-byte at boundary", IntType, thrItem, ""},
		{"non-byte just over boundary", IntType, thrItem + 1, "larger than address space"},
		{"non-byte maxint64", IntType, math.MaxInt64, "type [9223372036854775807]int larger than address space"},
		{"byte at boundary", Uint8Type, thrByte, ""},
		{"byte just over boundary", Uint8Type, thrByte + 1, "larger than address space"},
		{"byte maxint64", Uint8Type, math.MaxInt64, "type [9223372036854775807]uint8 larger than address space"},
		// 1<<62 overflows the non-byte (allocArrayItem/elem) accounting but
		// still fits the byte (1/elem) path: the two branches must diverge.
		{"cross-branch rejected as non-byte", IntType, 1 << 62, "larger than address space"},
		{"cross-branch accepted as byte", Uint8Type, 1 << 62, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMsg == "" {
				require.NotPanics(t, func() { checkArrayAllocFits(tt.et, tt.length) })
				return
			}
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected a panic")
				require.Contains(t, fmt.Sprint(r), tt.wantMsg)
			}()
			checkArrayAllocFits(tt.et, tt.length)
		})
	}
}

// TestEllipsisArrayIndexOverflow pins the variadic-array measurement guards
// that reject an [...]T literal whose implied length would overflow int64.
// This guard lives in preprocess1's measurement loop — a path the
// checkArrayAllocFits test above never reaches (it emits "array index ... out
// of bounds", not "larger than address space") — so it needs its own coverage.
// Mirrors the make23 (keyed) and make26 (unkeyed trailing element) filetests.
func TestEllipsisArrayIndexOverflow(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			// Keyed MaxInt64 implies length MaxInt64+1 (make23).
			"keyed maxint64",
			"[...]int{9223372036854775807: 1}",
			"array index 9223372036854775807 out of bounds",
		},
		{
			// Trailing unkeyed element after a MaxInt64-1 key reaches index
			// MaxInt64, overflowing the running length (make26).
			"unkeyed trailing element",
			"[...]int{9223372036854775806: 1, 2}",
			"array index 9223372036854775807 out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, _ := newPreprocessAllocTestStore(t, 64*1024*1024, stypes.NewInfiniteGasMeter())
			defer st.SetPreprocessAllocator(nil)

			pkgPath := "gno.land/r/test/ellipsis"
			m := NewMachineWithOptions(MachineOptions{
				PkgPath: pkgPath,
				Store:   st,
				Output:  io.Discard,
				Alloc:   NewAllocator(64 * 1024 * 1024),
			})
			defer m.Release()

			mpkg := &std.MemPackage{
				Type: MPUserProd,
				Name: "ellipsis",
				Path: pkgPath,
				Files: []*std.MemFile{{Name: "a.gno", Body: fmt.Sprintf(
					"package ellipsis\nfunc main() { _ = %s }\n", tt.expr)}},
			}
			panicked, val := runMemPackageRecover(m, mpkg)
			require.True(t, panicked, "expected preprocess to reject overflowing length")
			require.Contains(t, fmt.Sprint(val), tt.want, "got: %v", val)
		})
	}
}

// TestShiftAmountGasOverflow pins the clamp in doOpShl/doOpShlAssign. The
// shift-amount gas charge runs before shlAssign enforces maxBigintShift, so an
// unvalidated amount above ~2.4e17 used to wrap int64 negative and surface as
// "gas must not be negative" from ConsumeGas instead of the real
// "shift amount exceeds maximum" error. Every over-cap amount must report the
// cap, whatever its magnitude.
func TestShiftAmountGasOverflow(t *testing.T) {
	for _, shift := range []string{
		"10001",                // just over maxBigintShift
		"4000000000",           // > MaxInt32
		"300000000000000000",   // int64(x)*OpCPUSlopeBigIntShl overflows int64
		"18446744073709551615", // MaxUint64
	} {
		t.Run(shift, func(t *testing.T) {
			gm := stypes.NewGasMeter(math.MaxInt64)
			st, _ := newPreprocessAllocTestStore(t, 64*1024*1024, gm)
			t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
			m := NewMachineWithOptions(MachineOptions{
				PkgPath: "gno.land/r/test/shiftgas",
				Store:   st,
				Output:  io.Discard,
				Alloc:   NewAllocator(math.MaxInt64),
			})
			t.Cleanup(m.Release)
			panicked, val := runMemPackageRecover(m, &std.MemPackage{
				Type:  MPUserProd,
				Name:  "shiftgas",
				Path:  "gno.land/r/test/shiftgas",
				Files: []*std.MemFile{{Name: "a.gno", Body: "package shiftgas\n\nconst _ = 1 << " + shift + "\n"}},
			})
			require.True(t, panicked, "over-cap shift must be rejected")
			got := fmt.Sprint(val)
			require.Contains(t, got, "exceeds maximum",
				"must report the shift cap, got: %v", got)
			require.NotContains(t, got, "gas must not be negative",
				"gas charge overflowed instead of reporting the cap: %v", got)
		})
	}
}

// TestPreprocessGas_BignumLiterals pins the big-number charges at the layer
// where the O(n^2) work actually runs — constant folding during preprocess.
// Constant folding inherits the store's preprocess-allocator gas meter (same
// wiring as keeper AddPackage / withQueryEvalMachine), so a small finite meter
// here mirrors the real charging path while staying fast and deterministic.
func TestPreprocessGas_BignumLiterals(t *testing.T) {
	newMachine := func(t *testing.T, pkgName string, gm stypes.GasMeter) *Machine {
		t.Helper()
		st, _ := newPreprocessAllocTestStore(t, 256*1024*1024, gm)
		t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/" + pkgName,
			Store:   st,
			Output:  io.Discard,
			Alloc:   NewAllocator(math.MaxInt64),
		})
		t.Cleanup(m.Release)
		return m
	}
	run := func(t *testing.T, m *Machine, pkgName, body string) (bool, any) {
		t.Helper()
		return runMemPackageRecover(m, &std.MemPackage{
			Type:  MPUserProd,
			Name:  pkgName,
			Path:  "gno.land/r/test/" + pkgName,
			Files: []*std.MemFile{{Name: "a.gno", Body: body}},
		})
	}

	// Chained big-int shift: 1<<10000<<10000<<... grows the operand ~10000
	// bits per step (each shift is within maxBigintShift, so the per-shift cap
	// never trips). doOpShl now charges by operand bit-width, so the fold
	// exhausts a modest gas budget after a few hundred shifts. Without that
	// charge the 2000-shift chain bills only ~760K gas (per shift amount only)
	// and would NOT OOG here — so this also guards against regressing part 2.
	t.Run("chained_shift_oog", func(t *testing.T) {
		m := newMachine(t, "shlgas", stypes.NewGasMeter(10_000_000))
		body := "package shlgas\n\nconst _ = 1" + strings.Repeat(" << 10000", 2000) + "\n"
		panicked, val := run(t, m, "shlgas", body)
		require.True(t, panicked, "chained shift must exhaust gas during fold")
		require.Contains(t, strings.ToLower(fmt.Sprint(val)), "gas",
			"panic should be out-of-gas, got: %v", val)
	})

	// Literal parse: the quadratic charge exhausts a modest gas budget before
	// the O(n^2) parse runs, on both INT and FLOAT (a trailing ".0" must not
	// bypass it). ~200K digits is past the OOG threshold for both slopes at a
	// 1e7 gas budget.
	bigLit := strings.Repeat("1", 200_000)
	for _, tc := range []struct{ name, lit string }{
		{"int_literal_oog", bigLit},
		{"float_literal_oog", bigLit + ".0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newMachine(t, "litoog", stypes.NewGasMeter(10_000_000))
			body := "package litoog\n\nconst _ = " + tc.lit + "\n"
			panicked, val := run(t, m, "litoog", body)
			require.True(t, panicked, "huge literal must exhaust gas")
			require.Contains(t, strings.ToLower(fmt.Sprint(val)), "gas", "got: %v", val)
		})
	}

	// Control: a large literal that the removed cap would have rejected
	// (20000 digits) and a small shift chain both fold cleanly under a
	// realistic budget — the fix meters, it does not forbid.
	t.Run("within_limits_ok", func(t *testing.T) {
		m := newMachine(t, "okpkg", stypes.NewGasMeter(1_000_000_000))
		body := "package okpkg\n\nconst _ = " + strings.Repeat("9", 20_000) +
			"\nconst _ = 1 << 10000 << 10000\n"
		panicked, val := run(t, m, "okpkg", body)
		require.False(t, panicked, "within-limit literals/shifts must fold, got: %v", val)
	})
}

// TestBigLitParseChargeIsEvalInvariant guards a consensus hazard: the gas
// charged for parsing a numeric literal must not depend on how many times that
// AST node has already been evaluated.
//
// doOpEval used to write the blank-identifier-stripped text back into
// x.Value. Charging on the raw length while mutating the node meant the first
// evaluation billed the separators and every later one did not, so an
// underscored literal cost 24 gas once and 20 gas thereafter.
//
// That was latent, not reachable: preprocess replaces every *BasicLitExpr with
// a *ConstExpr (preprocess.go, TRANS_LEAVE) and Transcribe treats *ConstExpr as
// a leaf, so no literal parsed out of .gno source is ever evaluated twice. It
// is pinned anyway -- gas that varies with evaluation history is the same shape
// as the mem-package-cache gas fork, and only the absence of a second eval
// stands between the two.
func TestBigLitParseChargeIsEvalInvariant(t *testing.T) {
	t.Parallel()

	// Long enough that the (len/10)^2 floor does not swallow the difference,
	// and with enough separators to shift len/10.
	raw := "89_" + strings.Repeat("123456789_", 10) + "1234567890"
	require.NotEqual(t, len(raw), len(strings.ReplaceAll(raw, "_", "")),
		"test literal must actually contain separators")

	for _, tc := range []struct {
		name string
		kind Word
	}{
		{"int", INT},
		{"float", FLOAT},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value := raw
			if tc.kind == FLOAT {
				value = raw + ".5"
			}
			// One shared node, evaluated repeatedly -- the cached-AST case.
			x := &BasicLitExpr{Kind: tc.kind, Value: value}

			var charges []int64
			for range 3 {
				meter := stypes.NewGasMeter(1_000_000_000)
				m := NewMachineWithOptions(MachineOptions{
					PkgPath:  "test",
					GasMeter: meter,
				})
				before := meter.GasConsumed()
				m.PushExpr(x)
				m.doOpEval()
				charges = append(charges, meter.GasConsumed()-before)
				m.Release()
			}

			// Pin the formula, not just its stability: the keeper-side test
			// mirrors this arithmetic by hand, and a mirror can only ever
			// catch an over-charge. If this assertion moves, that mirror
			// (bigLitParseGas) and the ADR residual table must move with it.
			slope := int64(OpCPUSlopeBigIntSetString)
			if tc.kind == FLOAT {
				slope = OpCPUSlopeBigDecParse
			}
			d10 := int64(len(value)) / 10
			require.Equal(t, d10*d10*slope/10*GasFactorCPU, charges[0],
				"charge formula changed")

			require.Equal(t, charges[0], charges[1],
				"gas must not depend on evaluation count (got %v)", charges)
			require.Equal(t, charges[1], charges[2],
				"gas must not depend on evaluation count (got %v)", charges)
			require.Equal(t, value, x.Value,
				"doOpEval must not mutate the shared AST node")
		})
	}
}

// TestTxPathLiteralCap pins the tx-path ceiling on a numeric literal.
// TypeCheckMemPackage runs go/types before preprocess, and go/types refuses
// any numeric literal over 10000 characters, so no literal a transaction can
// deploy reaches chargeBigLitParse anywhere near the size at which the charge
// itself would refuse (~1.23M INT / ~866K FLOAT).
//
// That cap is go/types' own const (go/types/literals.go), not something the
// GoVersion pinned in gotypecheck.go governs, so a toolchain bump can move it
// silently. This test is what makes that visible.
func TestTxPathLiteralCap(t *testing.T) {
	t.Parallel()

	const goTypesLiteralCap = 10000

	typeCheckVar := func(t *testing.T, lit string) error {
		t.Helper()
		mpkg := &std.MemPackage{
			Type: MPUserProd,
			Name: "capprobe",
			Path: "gno.land/r/test/capprobe",
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: GenGnoModLatest("gno.land/r/test/capprobe")},
				{Name: "a.gno", Body: "package capprobe\n\nvar X = " + lit + "\n"},
			},
		}
		_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
			Getter: mockPackageGetter{},
			Mode:   TCLatestStrict,
		})
		return err
	}

	// Every shape the charge covers, at the size the ADR calls a tx-path
	// behaviour change.
	for _, tc := range []struct{ name, lit string }{
		{"decimal_int", strings.Repeat("9", 1_300_000)},
		{"hex_int", "0x" + strings.Repeat("f", 1_300_000)},
		{"frac_float", "1." + strings.Repeat("9", 1_299_998)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorContains(t, typeCheckVar(t, tc.lit), "excessively long constant")
		})
	}

	t.Run("boundary", func(t *testing.T) {
		t.Parallel()
		if atCap := typeCheckVar(t, strings.Repeat("9", goTypesLiteralCap)); atCap != nil {
			require.NotContains(t, atCap.Error(), "excessively long constant")
		}
		require.ErrorContains(t,
			typeCheckVar(t, strings.Repeat("9", goTypesLiteralCap+1)),
			"excessively long constant")
	})

	// Price the worst literal a transaction can carry through to doOpEval.
	t.Run("worst_tx_literal_gas", func(t *testing.T) {
		t.Parallel()
		d10 := int64(goTypesLiteralCap) / 10
		gas := d10 * d10 * OpCPUSlopeBigDecParse / 10 * GasFactorCPU
		require.Less(t, gas, int64(3_000_000_000)/1000)
	})
}
