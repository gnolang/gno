package calibrate

// Calibration benchmark for chain/reflect.objectInfo.
//
// Flat. The stamped path does four things, none of them scaling with anything
// the caller controls: hex-and-itoa the ObjectID, hash-and-bech32 an address
// from a fixed 20-byte realm hash plus a uint64, resolve the creating realm's
// path, and read the declared type's name. The unstamped and no-identity paths
// return earlier and cost strictly less, so the stamped case sets the price.
//
// The machine carries a real store holding a persisted realm, because the realm
// lookup is part of the measured work and a nil store would skip it. Base is a
// concrete heap item rather than a RefValue, so no object load happens in the
// timed loop; an object load is metered by the KVStore (gctx), not by this row.

import (
	"math"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

const reflectBenchPkgPath = "gno.land/r/x/reflectbench"

// newReflectObjectInfoBench builds the shape the native receives from Gno: a
// pointer to a standalone heap item, owned by a realm that exists in the store.
func newReflectObjectInfoBench(b *testing.B, stamped bool) *dispatchHarness {
	b.Helper()

	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	store := gno.NewStore(gno.NewAllocator(math.MaxInt64), baseStore, iavlStore)
	rlm := gno.NewRealm(reflectBenchPkgPath)
	rlm.Time = 1
	store.SetPackageRealm(rlm)
	tx := store.BeginTransaction(baseStore.CacheWrap(), iavlStore.CacheWrap(), nil, nil)
	// Load the realm into the transaction's cache, which is what production
	// looks like: the realm that created an object is loaded whenever that
	// object is reachable, so the native's realm lookup is a cache hit.
	tx.GetPackageRealm(reflectBenchPkgPath)

	alloc := gno.NewAllocator(math.MaxInt64)
	hiv := alloc.NewHeapItem(nil, gno.TypedValue{})
	hiv.SetPkgID(gno.PkgIDFromPkgPath(reflectBenchPkgPath))
	if stamped {
		hiv.SetNewTime(7)
	}

	m := newDispatchMachine(1)
	m.Store = tx
	m.Blocks[0].Values[0] = gno.TypedValue{V: gno.PointerValue{TV: &hiv.Value, Base: hiv}}

	return &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/reflect", "objectInfo"), nReturns: 6}
}

func BenchmarkNative_Reflect_ObjectInfo_Stamped(b *testing.B) {
	h := newReflectObjectInfoBench(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Reflect_ObjectInfo_Unstamped(b *testing.B) {
	h := newReflectObjectInfoBench(b, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// A value with no identity at all: the earliest return in the native, and the
// floor the flat row must still cover.
func BenchmarkNative_Reflect_ObjectInfo_NoIdentity(b *testing.B) {
	m := newDispatchMachine(1)
	m.Blocks[0].Values[0] = gno.TypedValue{T: gno.StringType}
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/reflect", "objectInfo"), nReturns: 6}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}
