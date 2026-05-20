// PKGPATH: gno.land/r/test/impl
package impl

// This test exercises getOwner returning nil for a deleted owner object.
//
// Tx1 (init): one entry "a" is appended to realm F's slice. This persists:
//   - an ArrayValue (cap 1) holding the slice's backing storage
//   - a HeapItemValue boxing *Entry "a", whose OwnerID points to that array
//
// Tx2 (main): append("c") grows the backing array. The cap-1 array from
// Tx1 is deleted in Tx2's finalization and replaced by a new cap-N array.
// The HeapItemValue from Tx1 is untouched on disk and still carries
// OwnerID = (now-deleted cap-1 array).
//
// During Tx2's finalization, markDirtyAncestors walks up from dirty
// objects and reaches that HeapItemValue. Following its OwnerID with
// the unsafe GetObject would panic — the object is gone. The fix is
// to call GetObjectSafe and treat a nil result as "stop the walk."
//
// Without this fix: panic: unexpected object with id ...:N
// With this fix:    walk stops at nil, test passes.

import (
	"gno.land/r/tests/vm/crossrealm_f"
)

func init() {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
}

func main(cur realm) {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "c", Value: 3})
	println(crossrealm_f.Len())
}


// Realm oplog legend (object IDs are from this run; package = crossrealm_f):
//   :2  PackageValue (escaped, RefCount=2)
//   :3  PackageBlock — holds the `entries` slice header
//   :9  OLD backing array (cap 1) from Tx1 — DELETED in Tx2 (see d[…:9])
//   :10 OLD HeapItemValue boxing *Entry "a" from Tx1 — its OwnerID still
//       points at :9. The u[…:10] update is markDirtyAncestors walking
//       through it; without GetObjectSafe, that walk would call
//       GetObject(:9) and panic because :9 is the d[…] line below.
//   :12 NEW backing array (cap 2) created by main's append
//   :13 NEW HeapItemValue boxing *Entry "c"
//   :14 NEW Entry{Key:"c", Value:3}
//
// The thing to notice is that d[…:9] (deleted) sits in the same
// finalization as u[…:10] (updated), and :10's OwnerID = :9. That
// adjacency is the bug-trigger: a dirty walk that follows :10 → :9.

// Error:
// cannot allocate gno.land/r/tests/vm/crossrealm_f.Entry in realm gno.land/r/test/impl
