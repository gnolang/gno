// PKGPATH: gno.land/r/test/impl
package impl

// This test exercises getOwner returning nil for a deleted owner object.
//
// During init(), two entries are added to realm D's slice via append().
// Each append may grow the backing array, causing the old ArrayValue to
// be saved then deleted within the same finalization. However, a
// HeapItemValue created during the same init still has its OwnerID
// pointing to the now-deleted ArrayValue.
//
// When main() runs and adds another entry, markDirtyAncestors walks up
// from dirty objects and reaches that HeapItemValue. Its OwnerID points
// to the deleted ArrayValue, which no longer exists in the store.
//
// getOwner must use GetObjectSafe and return nil to stop the walk,
// rather than panic with GetObject.
//
// Without this fix: panic: unexpected object with id ...:9
// With this fix:    walk stops at nil, test passes.

import (
	"gno.land/r/tests/vm/crossrealm_f"
)

func init() {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "b", Value: 2})
}

func main(cur realm) {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "c", Value: 3})
	println(crossrealm_f.Len())
}

// Output:
// 3
