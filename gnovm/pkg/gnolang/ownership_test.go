package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetOwnerDropsStoreFallback demonstrates bug H7:
// The old getOwner(store, oo) helper loaded the owner from the store when
// oo.GetOwner() returned nil but oo.OwnerID was set (common after
// deserialization / lazy loading). That helper was deleted; now only
// oo.GetOwner() is called, which returns nil when the in-memory owner
// pointer hasn't been set — even if OwnerID is non-zero.
//
// Additionally, GetOwnerID() returns ObjectID{} when owner is nil,
// ignoring the stored OwnerID field. This creates an inconsistency:
//
//	GetIsOwned()  -> true  (checks OwnerID, the stored field)
//	GetOwner()    -> nil   (checks owner, the memory pointer)
//	GetOwnerID()  -> zero  (checks owner, not OwnerID)
//
// In realm finalization (processNewEscapedMarks, markDirtyAncestors),
// the ownership chain walk terminates prematurely because GetOwner()
// returns nil for deserialized objects whose owners haven't been loaded.
func TestGetOwnerDropsStoreFallback(t *testing.T) {
	t.Parallel()

	// Simulate an object after deserialization: OwnerID is set from
	// stored data, but the in-memory owner pointer is nil.
	// This is exactly what ObjectInfo.Copy() produces (comment on line 181:
	// "Note that 'owner' is nil").
	ownerPkgID := PkgIDFromPkgPath("gno.land/r/demo/boards")
	ownerOID := ObjectID{PkgID: ownerPkgID, NewTime: 1}

	child := &StructValue{}
	child.ID = ObjectID{PkgID: ownerPkgID, NewTime: 2}
	child.OwnerID = ownerOID // set as deserialization would
	// child.owner is nil — not loaded from store yet

	// GetIsOwned checks the public OwnerID field — correctly reports owned.
	assert.True(t, child.GetIsOwned(),
		"GetIsOwned() should be true: OwnerID is set from deserialization")

	// BUG: GetOwner() returns nil because the in-memory pointer was never set.
	// The deleted getOwner(store, oo) helper would have loaded it from the store.
	assert.Nil(t, child.GetOwner(),
		"GetOwner() returns nil: in-memory owner pointer not set (no store fallback)")

	// BUG: GetOwnerID() also returns zero, because it checks owner (nil),
	// not the stored OwnerID field. This contradicts GetIsOwned().
	assert.True(t, child.GetOwnerID().IsZero(),
		"GetOwnerID() returns zero even though OwnerID is set — checks owner pointer, not OwnerID field")

	// The three methods are inconsistent:
	//   GetIsOwned()  = true   (from OwnerID)
	//   GetOwner()    = nil    (from owner)
	//   GetOwnerID()  = zero   (from owner)
	//
	// In realm finalization, code like:
	//   po := oo.GetOwner()
	//   if po == nil { continue }  // silently skips owned objects
	// will skip this object, breaking the ownership chain walk.
	assert.NotEqual(t, child.OwnerID.IsZero(), child.GetOwnerID().IsZero(),
		"OwnerID field and GetOwnerID() disagree — deserialized OwnerID is lost")
}
