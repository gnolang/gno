package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZerobaseStore_SentinelIdentity locks in the per-Store, per-TypeID
// sentinel identity: repeated Zerobase(t) calls return the same
// *HeapItemValue, distinct types return distinct instances, and the
// sentinel carries the reserved zerobase ObjectID.
func TestZerobaseStore_SentinelIdentity(t *testing.T) {
	st := newTestStore()

	emptyStruct := &StructType{Fields: nil}
	emptyArr := &ArrayType{Len: 0, Elt: IntType}

	hi1 := st.Zerobase(emptyStruct)
	hi2 := st.Zerobase(emptyStruct)
	require.True(t, hi1 == hi2, "same TypeID must return identical *HeapItemValue")
	require.True(t, hi1.GetObjectID().IsZerobase())
	require.Equal(t, zerobaseObjectID, hi1.GetObjectID())

	hiArr := st.Zerobase(emptyArr)
	require.False(t, hi1 == hiArr, "distinct TypeIDs must return distinct *HeapItemValue")
	require.True(t, hiArr.GetObjectID().IsZerobase())

	// HIV Value carries the proper element type so dereferencing /
	// printing through reflect sees the right type tag.
	assert.Equal(t, emptyStruct.TypeID(), hi1.Value.T.TypeID())
	assert.Equal(t, emptyArr.TypeID(), hiArr.Value.T.TypeID())
}

// TestZerobaseStore_FillFromRefValue locks in the values.go PointerValue
// load path: when a persisted PointerValue arrives with Base =
// RefValue{ObjectID: zerobaseObjectID} (and nil TV), fillValueTV must
// rebind Base and TV to the in-memory per-type sentinel recovered via
// Store.Zerobase(elt) — using the outer PointerType to identify which
// sentinel to fetch. Without this branch, fillValueTV would fall through
// to store.GetObject(sentinelOID) and trip the invariant panic.
func TestZerobaseStore_FillFromRefValue(t *testing.T) {
	st := newTestStore()
	emptyStruct := &StructType{Fields: nil}
	sentinel := st.Zerobase(emptyStruct)

	// Construct a TypedValue mimicking the shape amino-decoding leaves:
	// PointerValue with RefValue base, TV stripped, outer type *struct{}.
	tv := &TypedValue{
		T: &PointerType{Elt: emptyStruct},
		V: PointerValue{
			Base:  RefValue{ObjectID: zerobaseObjectID},
			Index: 0,
		},
	}

	fillValueTV(st, tv)

	pv := tv.V.(PointerValue)
	require.True(t, pv.Base.(*HeapItemValue) == sentinel,
		"Base must be rebound to the in-memory sentinel HIV")
	require.True(t, pv.TV == &sentinel.Value,
		"TV must be rebound to &sentinel.Value so == on PointerValue holds")
}

// TestZerobaseStore_FillRoutesByElementType ensures the load path
// recovers each pointer's element type from the outer PointerType and
// routes to the matching per-TypeID sentinel — no cross-wiring between
// types that share the same reserved ObjectID.
func TestZerobaseStore_FillRoutesByElementType(t *testing.T) {
	st := newTestStore()
	emptyStruct := &StructType{Fields: nil}
	emptyArr := &ArrayType{Len: 0, Elt: IntType}

	sentinelStruct := st.Zerobase(emptyStruct)
	sentinelArr := st.Zerobase(emptyArr)
	require.False(t, sentinelStruct == sentinelArr)

	// Identical RefValue base — both reference the reserved zerobase OID.
	bareRef := RefValue{ObjectID: zerobaseObjectID}

	tvStruct := &TypedValue{
		T: &PointerType{Elt: emptyStruct},
		V: PointerValue{Base: bareRef, Index: 0},
	}
	tvArr := &TypedValue{
		T: &PointerType{Elt: emptyArr},
		V: PointerValue{Base: bareRef, Index: 0},
	}
	fillValueTV(st, tvStruct)
	fillValueTV(st, tvArr)

	require.True(t, tvStruct.V.(PointerValue).Base.(*HeapItemValue) == sentinelStruct,
		"*struct{} ref must resolve to the struct sentinel")
	require.True(t, tvArr.V.(PointerValue).Base.(*HeapItemValue) == sentinelArr,
		"*[0]int ref must resolve to the array sentinel")
}

// TestZerobaseStore_GetObjectInvariantPanics verifies that any direct
// GetObject lookup of a sentinel ObjectID panics with the documented
// invariant message. The load path's special-case is supposed to keep
// sentinel OIDs out of GetObject entirely.
func TestZerobaseStore_GetObjectInvariantPanics(t *testing.T) {
	st := newTestStore()
	assert.Panics(t, func() {
		st.GetObject(zerobaseObjectID)
	}, "GetObject on a sentinel OID must trip the invariant panic")
}

func newTestStore() *defaultStore {
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	return NewStore(NewAllocator(1<<30), tm2, tm2)
}
