package calibrate

// Calibration benchmark for chain/reflect.objectID.
//
// Flat: one pointer-base resolution plus, for a stamped object, the ObjectID's
// hex+itoa spelling. The ID is a fixed 20-byte realm hash and a uint64, so
// there is no caller-controlled length term and no slope. The unstamped and
// no-identity paths return earlier and cost strictly less, so the stamped case
// sets the price.
//
// The harness builds the shape the native actually receives from Gno: a
// pointer to a standalone heap item. Base is the concrete heap item rather
// than a RefValue, so no store read happens inside the timed loop: a store
// read is metered by the KVStore (gctx), not by the native's own row.

import (
	"math"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func newReflectObjectIDBench(b *testing.B, stamped bool) *dispatchHarness {
	b.Helper()
	alloc := gno.NewAllocator(math.MaxInt64)
	hiv := alloc.NewHeapItem(nil, gno.TypedValue{})
	hiv.SetPkgID(gno.PkgIDFromPkgPath("gno.land/r/x"))
	if stamped {
		hiv.SetNewTime(7)
	}

	m := newDispatchMachine(1)
	m.Blocks[0].Values[0] = gno.TypedValue{V: gno.PointerValue{TV: &hiv.Value, Base: hiv}}
	return &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/reflect", "objectID"), nReturns: 3}
}

func BenchmarkNative_Reflect_ObjectID_Stamped(b *testing.B) {
	h := newReflectObjectIDBench(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Reflect_ObjectID_Unstamped(b *testing.B) {
	h := newReflectObjectIDBench(b, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// A value with no identity at all: the earliest return in the native, and the
// floor the flat row must still cover.
func BenchmarkNative_Reflect_ObjectID_NoIdentity(b *testing.B) {
	m := newDispatchMachine(1)
	m.Blocks[0].Values[0] = gno.TypedValue{T: gno.StringType}
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/reflect", "objectID"), nReturns: 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}
