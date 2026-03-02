package gnolang

// XXX test that p is not actually mutable

// XXX finalize should consider hard boundaries only

// XXX types: to support realm persistence of types, must
// first require the validation of blocknode locations.

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
)

/*
## Realms

Gno is designed with blockchain smart contract programming in
mind.  A smart-contract enabled blockchain is like a
massive-multiuser-online operating-system (MMO-OS). Each user
is provided a home package, for example
"gno.land/r/username". This is not just a regular package but
a "realm package", and functions and methods declared there
have special privileges.

Every "realm package" should define at least one package-level variable:

```go
// PKGPATH: gno.land/r/alice
package alice
var root interface{}

func UpdateRoot(...) error {
  root = ...
}
```

Here, the root variable can be any object, and indicates the
root node in the data realm identified by the package path
"gno.land/r/alice".

Any number of package-level values may be declared in a
realm; they are all owned by the package and get
merkle-hashed into a single root hash for the package realm.

The gas cost of transactions that modify state are paid for
by whoever submits the transaction, but the storage rent is
paid for by the realm.  Anyone can pay the storage upkeep of
a realm to keep it alive.
*/

//----------------------------------------
// PkgID & Realm

type PkgID struct {
	Hashlet
}

func (pid PkgID) MarshalAmino() (string, error) {
	return hex.EncodeToString(pid.Hashlet[:]), nil
}

func (pid *PkgID) UnmarshalAmino(h string) error {
	_, err := hex.Decode(pid.Hashlet[:], []byte(h))
	return err
}

func (pid PkgID) String() string {
	return fmt.Sprintf("RID%X", pid.Hashlet[:])
}

func (pid PkgID) Bytes() []byte {
	return pid.Hashlet[:]
}

var (
	pkgIDFromPkgPathCacheMu sync.Mutex // protects the shared cache.
	// TODO: later on switch this to an LRU if needed to ensure
	// fixed memory caps. For now though it isn't a problem:
	// https://github.com/gnolang/gno/pull/3424#issuecomment-2564571785
	pkgIDFromPkgPathCache = make(map[string]*PkgID, 100)
)

func PkgIDFromPkgPath(path string) PkgID {
	pkgIDFromPkgPathCacheMu.Lock()
	defer pkgIDFromPkgPathCacheMu.Unlock()

	pkgID, ok := pkgIDFromPkgPathCache[path]
	if !ok {
		pkgID = new(PkgID)
		*pkgID = PkgID{HashBytes([]byte(path))}
		pkgIDFromPkgPathCache[path] = pkgID
	}
	return *pkgID
}

// Returns the ObjectID of the PackageValue associated with path.
func ObjectIDFromPkgPath(path string) ObjectID {
	pkgID := PkgIDFromPkgPath(path)
	return ObjectIDFromPkgID(pkgID)
}

// Returns the ObjectID of the PackageValue associated with pkgID.
func ObjectIDFromPkgID(pkgID PkgID) ObjectID {
	return ObjectID{
		PkgID:   pkgID,
		NewTime: 1, // by realm logic.
	}
}

// --------------------------------------------------------------------------------
// Realm
var nilRealm = (*Realm)(nil)

// NOTE: A nil realm is special and has limited functionality; enough to
// support methods that don't require persistence. This is the default realm
// when a machine starts with a non-realm package.
type Realm struct {
	ID   PkgID
	Path string
	Time uint64

	Deposit uint64 // Amount of deposit held
	Storage uint64 // Amount of storage used
	sumDiff int64  // Total size difference from added, updated, or deleted objects

	newCreated []Object
	newDeleted []Object
	newEscaped []Object

	created []Object // about to become real.
	updated []Object // real objects that were modified.
	deleted []Object // real objects that became deleted.
	escaped []Object // real objects with refcount > 1.
}

// Creates a blank new realm with counter 0.
func NewRealm(path string) *Realm {
	id := PkgIDFromPkgPath(path)
	return &Realm{
		ID:   id,
		Path: path,
		Time: 0,
	}
}

func (rlm *Realm) GetPath() string {
	if rlm == nil {
		return ""
	} else {
		return rlm.Path
	}
}

func (rlm *Realm) String() string {
	if rlm == nil {
		return "Realm(nil)"
	} else {
		return fmt.Sprintf(
			"Realm{Path:%q,Time:%d}#%X",
			rlm.Path, rlm.Time, rlm.ID.Bytes())
	}
}

//----------------------------------------
// ownership hooks

// po's old elem value is xo, will become co.
// po, xo, and co may each be nil.
// if rlm or po is nil, do nothing.
// xo or co is nil if the element value is undefined or has no
// associated object.
func (rlm *Realm) DidUpdate(po, xo, co Object) {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
	if rlm == nil {
		return
	}
	if debugRealm {
		if co != nil && co.GetIsDeleted() {
			panic("cannot attach a deleted object")
		}
		if po != nil && po.GetIsTransient() {
			panic("cannot attach to a transient object")
		}
		if po != nil && po.GetIsDeleted() {
			panic("cannot attach to a deleted object")
		}
	}
	if po == nil || !po.GetIsReal() {
		return // do nothing.
	}
	if po.GetObjectID().PkgID != rlm.ID {
		panic(&Exception{Value: typedString("cannot modify external-realm or non-realm object")})
	}

	// XXX check if this boosts performance
	// XXX with broad integration benchmarking.
	// XXX if co == xo {
	// XXX }

	// From here on, po is real (not new-real).
	// Updates to .newCreated/.newEscaped /.newDeleted made here. (first gen)
	// More appends happen during FinalizeRealmTransactions(). (second+ gen)
	rlm.MarkDirty(po)

	if co != nil {
		co.IncRefCount()
		if co.GetRefCount() > 1 {
			if !co.GetIsEscaped() {
				rlm.MarkNewEscaped(co)
			}
		}
		if co.GetIsReal() {
			rlm.MarkDirty(co)
		} else {
			co.SetOwner(po)
			rlm.MarkNewReal(co)
		}
	}

	if xo != nil {
		xo.DecRefCount()
		if xo.GetRefCount() == 0 {
			if xo.GetIsReal() {
				rlm.MarkNewDeleted(xo)
			}
		} else if xo.GetIsReal() {
			rlm.MarkDirty(xo)
		}
	}
}

//----------------------------------------
// mark*

func (rlm *Realm) MarkNewReal(oo Object) {
	if debugRealm {
		if pv, ok := oo.(*PackageValue); ok {
			// packages should have no owner.
			if pv.GetOwner() != nil {
				panic("cannot mark owned package as new real")
			}
			// packages should have ref-count 1.
			if pv.GetRefCount() != 1 {
				panic("cannot mark non-singly referenced package as new real")
			}
		} else {
			if oo.GetOwner() == nil {
				panic("cannot mark unowned object as new real")
			}
			if !oo.GetOwner().GetIsReal() {
				panic("cannot mark object as new real if owner is not real")
			}
		}
	}
	if oo.GetIsNewReal() {
		return // already marked.
	}
	oo.SetIsNewReal(true)
	// append to .newCreated
	if rlm.newCreated == nil {
		rlm.newCreated = make([]Object, 0, 256)
	}
	rlm.newCreated = append(rlm.newCreated, oo)
}

func (rlm *Realm) MarkDirty(oo Object) {
	if debugRealm {
		if !oo.GetIsReal() && !oo.GetIsNewReal() {
			panic("cannot mark unreal object as dirty")
		}
	}
	if oo.GetIsDirty() {
		return // already marked.
	}
	if oo.GetIsNewReal() {
		return // treat as new-real.
	}
	oo.SetIsDirty(true, rlm.Time)
	// append to .updated
	if rlm.updated == nil {
		rlm.updated = make([]Object, 0, 256)
	}
	rlm.updated = append(rlm.updated, oo)
}

func (rlm *Realm) MarkNewDeleted(oo Object) {
	if debugRealm {
		if !oo.GetIsNewReal() && !oo.GetIsReal() {
			panic("cannot mark unreal object as new deleted")
		}
		if oo.GetIsDeleted() {
			panic("cannot mark deleted object as new deleted")
		}
	}
	if oo.GetIsNewDeleted() {
		return // already marked.
	}
	oo.SetIsNewDeleted(true)
	// append to .newDeleted
	if rlm.newDeleted == nil {
		rlm.newDeleted = make([]Object, 0, 256)
	}
	rlm.newDeleted = append(rlm.newDeleted, oo)
}

func (rlm *Realm) MarkNewEscaped(oo Object) {
	if debugRealm {
		if !oo.GetIsNewReal() && !oo.GetIsReal() {
			panic("cannot mark unreal object as new escaped")
		}
		if oo.GetIsDeleted() {
			panic("cannot mark deleted object as new escaped")
		}
		if oo.GetIsEscaped() {
			panic("cannot mark escaped object as new escaped")
		}
	}
	if oo.GetIsNewEscaped() {
		return // already marked.
	}
	oo.SetIsNewEscaped(true)
	// append to .newEscaped.
	if rlm.newEscaped == nil {
		rlm.newEscaped = make([]Object, 0, 256)
	}
	rlm.newEscaped = append(rlm.newEscaped, oo)
}

//----------------------------------------
// transactions

// OpReturn calls this when exiting a realm transaction.
func (rlm *Realm) FinalizeRealmTransaction(store Store) {
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}

	if debugRealm {
		// * newCreated - may become created unless ancestor is deleted
		// * newDeleted - may become deleted unless attached to new-real owner
		// * newEscaped - may become escaped unless new-real and refcount 0 or 1.
		// * updated - includes all real updated objects, and will be appended with ancestors
		ensureUniq(rlm.newCreated)
		ensureUniq(rlm.newDeleted)
		ensureUniq(rlm.newEscaped)
		ensureUniq(rlm.updated)
		if false ||
			rlm.created != nil ||
			rlm.deleted != nil ||
			rlm.escaped != nil {
			panic("realm should not have created, deleted, or escaped marks before beginning finalization")
		}
	}
	// log realm boundaries in opslog.
	store.LogFinalizeRealm(rlm.Path)
	// increment recursively for created descendants.
	// also assigns object ids for all.
	rlm.processNewCreatedMarks(store, 0)
	// decrement recursively for deleted descendants.
	rlm.processNewDeletedMarks(store)
	// at this point, all ref-counts are final.
	// demote any escaped if ref-count is 1.
	rlm.processNewEscapedMarks(store, 0)
	// given created and updated objects,
	// mark all owned-ancestors also as dirty.
	rlm.markDirtyAncestors(store)
	if debugRealm {
		ensureUniq(rlm.created, rlm.updated)
		ensureUniq(rlm.escaped)
	}
	// save all the created and updated objects.
	// hash calculation is done along the way,
	// or via escaped-object persistence in
	// the iavl tree.
	rlm.saveUnsavedObjects(store)
	// TODO saved newly escaped to iavl.
	rlm.saveNewEscaped(store)
	// delete all deleted objects.
	rlm.removeDeletedObjects(store)
	// reset realm state for new transaction.
	rlm.clearMarks()

	// Update storage differences.
	realmDiffs := store.RealmStorageDiffs()
	realmDiffs[rlm.Path] += rlm.sumDiff
	rlm.sumDiff = 0
}

//----------------------------------------
// processNewCreatedMarks

// Crawls marked created children and increments ref counts,
// finding more newly created objects recursively.
// All newly created objects become appended to .created,
// and get assigned ids.
// Starts processing with index 'start', returns len(newCreated).
func (rlm *Realm) processNewCreatedMarks(store Store, start int) int {
	// Create new objects and their new descendants.
	for _, oo := range rlm.newCreated[start:] {
		if debugRealm {
			if oo.GetIsDirty() {
				panic("new created mark cannot be dirty")
			}
		}
		if oo.GetRefCount() == 0 {
			if debugRealm {
				// The refCount for a new real object could be zero,
				// and the object may not yet be marked as deleted.
				if !oo.GetIsNewDeleted() && !oo.GetIsNewReal() {
					panic("should have been marked new-deleted")
				}
			}
			// No need to unmark, will be garbage collected.
			// oo.SetIsNewReal(false)
			// skip if became deleted.
			continue
		} else {
			rlm.incRefCreatedDescendants(store, oo)
		}
	}
	// Save new realm time.
	if len(rlm.newCreated) > 0 {
		store.SetPackageRealm(rlm)
	}
	return len(rlm.newCreated)
}

// oo must be marked new-real, and ref-count already incremented.
func (rlm *Realm) incRefCreatedDescendants(store Store, oo Object) {
	if debugRealm {
		if oo.GetIsDirty() {
			panic("cannot increase reference of descendants of dirty objects")
		}
		if oo.GetRefCount() <= 0 {
			panic("cannot increase reference of descendants of unreferenced object")
		}
	}

	// RECURSE GUARD
	// if id already set, skip.
	// this happens when a node marked created was already
	// visited via recursion from a prior marked created.
	if !oo.GetObjectID().IsZero() {
		return
	}
	rlm.assignNewObjectID(oo)
	rlm.created = append(rlm.created, oo)
	// RECURSE GUARD END

	// recurse for children.
	more := getChildObjects2(store, oo)
	for _, child := range more {
		if _, ok := child.(*PackageValue); ok {
			if debugRealm {
				if child.GetRefCount() < 1 {
					panic("cannot increase reference count of package descendant that is unreferenced")
				}
			}
			// extern package values are skipped.
			continue
		}
		child.IncRefCount()
		rc := child.GetRefCount()
		if rc == 1 {
			if child.GetIsReal() {
				// a deleted real became undeleted.
				child.SetOwner(oo)
				rlm.MarkDirty(child)
			} else {
				// a (possibly pre-existing) new object
				// became real (again).
				// NOTE: may already be marked for first gen
				// newCreated or updated.
				child.SetOwner(oo)

				// Mark it as new-real first to prevent it
				// from being marked dirty upon reentry.
				child.SetIsNewReal(true)
				rlm.incRefCreatedDescendants(store, child)
			}
		} else if rc > 1 {
			// new real or dirty shouldn't be marked.
			rlm.MarkDirty(child)
			if child.GetIsEscaped() {
				// already escaped, do nothing.
			} else {
				// NOTE: do not unset owner here,
				// may become unescaped later
				// in processNewEscapedMarks().
				// NOTE: may already be escaped.
				rlm.MarkNewEscaped(child)
			}
		} else {
			panic("child reference count should be greater than zero after increasing")
		}
	}
}

//----------------------------------------
// processNewDeletedMarks

// Crawls marked deleted children and decrements ref counts,
// finding more newly deleted objects, recursively.
// Recursively found deleted objects are appended
// to rlm.deleted.
// Must run *after* processNewCreatedMarks().
func (rlm *Realm) processNewDeletedMarks(store Store) {
	for _, oo := range rlm.newDeleted {
		if debugRealm {
			if oo.GetObjectID().IsZero() {
				panic("new deleted mark should have an object ID")
			}
		}
		if oo.GetRefCount() > 0 {
			oo.SetIsNewDeleted(false)
			// skip if became undeleted.
			continue
		} else {
			rlm.decRefDeletedDescendants(store, oo)
		}
	}
}

// Like incRefCreatedDescendants but decrements.
func (rlm *Realm) decRefDeletedDescendants(store Store, oo Object) {
	if debugRealm {
		if oo.GetObjectID().IsZero() {
			panic("cannot decrement references of deleted descendants of object with no object ID")
		}
		if oo.GetRefCount() != 0 {
			panic("cannot decrement references of deleted descendants of object with references")
		}
	}

	// RECURSE GUARD
	// if already deleted, skip.
	// this happens when a node marked deleted was already
	// deleted via recursion from a prior marked deleted.
	if oo.GetIsDeleted() {
		return
	}
	oo.SetIsNewDeleted(false)
	oo.SetIsNewReal(false)
	oo.SetIsNewEscaped(false)
	oo.SetIsDeleted(true, rlm.Time)
	rlm.deleted = append(rlm.deleted, oo)
	// RECURSE GUARD END

	// recurse for children
	more := getChildObjects2(store, oo)
	for _, child := range more {
		child.DecRefCount()
		rc := child.GetRefCount()
		if rc == 0 {
			rlm.decRefDeletedDescendants(store, child)
		} else if rc > 0 {
			rlm.MarkDirty(child)
		} else {
			panic("deleted descendants should not have a reference count of less than zero")
		}
	}
}

//----------------------------------------
// processNewEscapedMarks

// demotes new-real escaped objects with refcount 0 or 1.  remaining
// objects get their original owners marked dirty (to be further
// marked via markDirtyAncestors).
// Starts processing with index 'start', returns len(newEscaped).
func (rlm *Realm) processNewEscapedMarks(store Store, start int) int {
	escaped := make([]Object, 0, len(rlm.newEscaped))
	// These are those marked by MarkNewEscaped(),
	// regardless of whether new-real or was real,
	// but is always newly escaped,
	// (and never can be unescaped,)
	// except for new-reals that get demoted
	// because ref-count isn't >= 2.
	// for _, eo := range rlm.newEscaped[start:] {
	for i := 0; i < len(rlm.newEscaped[start:]); i++ { // may expand.
		eo := rlm.newEscaped[i]
		if debugRealm {
			if !eo.GetIsNewEscaped() {
				panic("new escaped mark not marked as new escaped")
			}
			if eo.GetIsEscaped() {
				panic("new escaped mark already escaped")
			}
		}
		if eo.GetRefCount() <= 1 {
			// demote; do not add to escaped.
			eo.SetIsNewEscaped(false)
			continue
		} else {
			// escape;
			// NOTE: do not unset new-escaped,
			// we do that upon actually persisting
			// the hash index.
			// eo.SetIsNewEscaped(false)
			escaped = append(escaped, eo)

			// add to escaped, and mark dirty previous owner.
			po := getOwner(store, eo)
			if po == nil {
				// e.g. !eo.GetIsNewReal(),
				// should have no parent.
				continue
			} else {
				if po.GetRefCount() == 0 {
					// is deleted, ignore.
				} else if po.GetIsNewReal() {
					// will be saved regardless.
				} else {
					// exists, mark dirty.
					rlm.MarkDirty(po)
				}
				if eo.GetObjectID().IsZero() {
					// eo was passed from caller.
					rlm.incRefCreatedDescendants(store, eo)
					eo.SetIsNewReal(true)
				}
				// escaped has no owner.
				eo.SetOwner(nil)
			}
		}
	}
	rlm.escaped = escaped
	return len(rlm.newEscaped)
}

//----------------------------------------
// markDirtyAncestors

// New and modified objects' owners and their owners
// (ancestors) must be marked as dirty to update the
// hash tree.
func (rlm *Realm) markDirtyAncestors(store Store) {
	markAncestors := func(oo Object) {
		for {
			if pv, ok := oo.(*PackageValue); ok {
				if debugRealm {
					if pv.GetRefCount() < 1 {
						panic("expected package value to have refcount 1 or greater")
					}
				}
				// package values have no ancestors.
				break
			}
			rc := oo.GetRefCount()
			if debugRealm {
				if rc == 0 {
					panic("ancestor should have a non-zero reference count to be marked as dirty")
				}
			}
			if rc > 1 {
				if debugRealm {
					if !oo.GetIsEscaped() && !oo.GetIsNewEscaped() {
						panic("ancestor should cannot be escaped or new escaped to be marked as dirty")
					}
					if !oo.GetOwnerID().IsZero() {
						panic("ancestor's owner ID cannot be zero to be marked as dirty")
					}
				}
				// object is escaped, so
				// it has no parent.
				break
			} // else, rc == 1

			po := getOwner(store, oo)

			if po == nil {
				break // no more owners.
			} else if po.GetIsNewReal() {
				// already will be marked
				// via call to markAncestors
				// via .created.
				break
			} else if po.GetIsDirty() {
				// already will be marked
				// via call to markAncestors
				// via .updated.
				break
			} else if po.GetIsDeleted() {
				// already deleted, no need to mark.
				// oo(child) maybe have another owner,
				// if so, it should be marked by .updated.
				break
			} else {
				rlm.MarkDirty(po)
				// next case
				oo = po
			}
		}
	}
	// NOTE: newly dirty-marked owners get appended
	// to .updated without affecting iteration.
	for _, oo := range rlm.updated {
		if !oo.GetIsDeleted() {
			markAncestors(oo)
		}
	}
	// NOTE: must happen after iterating over rlm.updated
	// for the same reason.
	for _, oo := range rlm.created {
		if !oo.GetIsDeleted() {
			markAncestors(oo)
		}
	}
}

//----------------------------------------
// saveUnsavedObjects

// Saves .created and .updated objects.
func (rlm *Realm) saveUnsavedObjects(store Store) {
	tids := make(map[TypeID]struct{})
	for _, co := range rlm.created {
		// for i := len(rlm.created) - 1; i >= 0; i-- {
		// co := rlm.created[i]
		if !co.GetIsNewReal() {
			// might have happened already as child
			// of something else created.
			continue
		} else {
			if !co.GetIsDeleted() {
				rlm.saveUnsavedObjectRecursively(store, co, tids)
			}
		}
	}
	for _, uo := range rlm.updated {
		// for i := len(rlm.updated) - 1; i >= 0; i-- {
		// uo := rlm.updated[i]
		if !uo.GetIsDirty() {
			// might have happened already as child
			// of something else created/dirty.
			continue
		} else {
			if !uo.GetIsDeleted() {
				// No recursive save needed; child objects were already
				// persisted via created objects.
				rlm.assertObjectIsPublic(uo, store, tids)
				rlm.saveObject(store, uo)
				uo.SetIsDirty(false, 0)
			}
		}
	}
}

// store unsaved children first.
// ensure that the object and children does not have any private dependencies.
// use a visited map to mark visited types when asserting there are no private dependencies.
func (rlm *Realm) saveUnsavedObjectRecursively(store Store, oo Object, visited map[TypeID]struct{}) {
	if debugRealm {
		if !oo.GetIsNewReal() && !oo.GetIsDirty() {
			panic("cannot save new real or non-dirty objects")
		}
		// object id should have been assigned during processNewCreatedMarks.
		if oo.GetObjectID().IsZero() {
			panic("cannot save object with no ID")
		}
		// deleted objects should not have gotten here.
		if false ||
			oo.GetRefCount() <= 0 ||
			oo.GetIsNewDeleted() ||
			oo.GetIsDeleted() {
			panic("cannot save deleted objects")
		}
	}

	// assert object have no private dependencies.
	//
	// XXX JAE: Can't this whole routine be changed so that it only applies
	// when finalizing when the first frame function is declared in a
	// private package? See discussion:
	// https://github.com/gnolang/gno/pull/4890/files#r2554336836
	rlm.assertObjectIsPublic(oo, store, visited)

	// first, save unsaved children.
	unsaved := getUnsavedChildObjects(oo)
	for _, uch := range unsaved {
		if uch.GetIsEscaped() || uch.GetIsNewEscaped() {
			// no need to save preemptively.
		} else {
			rlm.saveUnsavedObjectRecursively(store, uch, visited)
		}
	}
	// then, save self.
	if oo.GetIsNewReal() {
		// save created object.
		if debugRealm {
			if oo.GetIsDirty() {
				panic("cannot save dirty new real object")
			}
		}
		rlm.saveObject(store, oo)
		oo.SetIsNewReal(false)
	} else {
		// update existing object.
		if debugRealm {
			if !oo.GetIsDirty() {
				panic("cannot save non-dirty existing object")
			}
			if !oo.GetIsReal() {
				panic("cannot save unreal existing object")
			}
			if oo.GetIsNewReal() {
				panic("cannot save new real existing object")
			}
		}
		rlm.saveObject(store, oo)
		oo.SetIsDirty(false, 0)
	}
}

func (rlm *Realm) saveObject(store Store, oo Object) {
	oid := oo.GetObjectID()
	if oid.IsZero() {
		panic("unexpected zero object id")
	}
	// set hash to escape index.
	if oo.GetIsNewEscaped() {
		oo.SetIsNewEscaped(false)
		oo.SetIsEscaped(true)
		// XXX anything else to do?
	}

	// set object to store.
	// NOTE: also sets the hash to object.
	rlm.sumDiff += store.SetObject(oo)
}

//----------------------------------------
// saveNewEscaped

// Save newly escaped oid->hash to iavl for iavl proofs.
// NOTE some of these escaped items may be in external realms.
// TODO actually implement.
func (rlm *Realm) saveNewEscaped(store Store) {
	// TODO implement.
	/*
		for _, eo := range rlm.escaped {
				if !oo.GetIsEscaped() {
					panic("should not happen")
				}
		}
	*/
}

//----------------------------------------
// removeDeletedObjects

func (rlm *Realm) removeDeletedObjects(store Store) {
	for _, do := range rlm.deleted {
		rlm.sumDiff -= store.DelObject(do)
	}
}

//----------------------------------------
// clearMarks

func (rlm *Realm) clearMarks() {
	// sanity check
	if debugRealm {
		for _, oo := range rlm.newDeleted {
			if oo.GetIsNewDeleted() {
				panic("cannot clear marks if new deleted exist")
			}
		}

		// A new real object can be possible here.
		// This new real object may have recCount of 0
		// but its state was not unset. see `processNewCreatedMarks`.
		// (As a result, it will be garbage collected.)
		// therefore, new real is allowed to exist here.
		for _, oo := range rlm.newCreated {
			if oo.GetIsNewReal() && oo.GetRefCount() != 0 {
				panic("cannot clear marks if new created exist with refCount not zero")
			}
		}

		for _, oo := range rlm.newEscaped {
			if oo.GetIsNewEscaped() {
				panic("cannot clear marks if new escaped exist")
			}
		}
	}
	// reset
	rlm.newCreated = nil
	rlm.newEscaped = nil
	rlm.newDeleted = nil
	rlm.created = nil
	rlm.updated = nil
	rlm.deleted = nil
	rlm.escaped = nil
}

// assertObjectIsPublic ensures that the object is public and does not have any private dependencies.
// it check recursively the types of the object
// it does not recursively check the values because
// child objects are validated separately during the save traversal (saveUnsavedObjectRecursively)
func (rlm *Realm) assertObjectIsPublic(obj Object, store Store, visited map[TypeID]struct{}) {
	objID := obj.GetObjectID()
	if objID.PkgID != rlm.ID && isPkgPrivateFromPkgID(store, objID.PkgID) {
		panic("cannot persist reference of object from the private realm " + pkgPathFromPkgID(store, objID.PkgID))
	}

	// NOTE: should i set the visited tids map at the higher level so it's set one time.
	// it could help to reduce the number of checks for the same type.
	switch v := obj.(type) {
	case *HeapItemValue:
		if v.Value.T != nil {
			rlm.assertTypeIsPublic(store, v.Value.T, visited)
		}
	case *ArrayValue:
		for _, av := range v.List {
			if av.T != nil {
				rlm.assertTypeIsPublic(store, av.T, visited)
			}
		}
	case *StructValue:
		for _, sv := range v.Fields {
			if sv.T != nil {
				rlm.assertTypeIsPublic(store, sv.T, visited)
			}
		}
	case *MapValue:
		for head := v.List.Head; head != nil; head = head.Next {
			if head.Key.T != nil {
				rlm.assertTypeIsPublic(store, head.Key.T, visited)
			}
			if head.Value.T != nil {
				rlm.assertTypeIsPublic(store, head.Value.T, visited)
			}
		}
	case *FuncValue:
		if v.PkgPath != rlm.Path && isPkgPrivateFromPkgPath(store, v.PkgPath) {
			panic("cannot persist function or method from the private realm " + v.PkgPath)
		}
		if v.Type != nil {
			rlm.assertTypeIsPublic(store, v.Type, visited)
		}
		for _, capture := range v.Captures {
			if capture.T != nil {
				rlm.assertTypeIsPublic(store, capture.T, visited)
			}
		}
	case *BoundMethodValue:
		if v.Func.PkgPath != rlm.Path && isPkgPrivateFromPkgPath(store, v.Func.PkgPath) {
			panic("cannot persist bound method from the private realm " + v.Func.PkgPath)
		}
		if v.Receiver.T != nil {
			rlm.assertTypeIsPublic(store, v.Receiver.T, visited)
		}
	case *Block:
		for _, bv := range v.Values {
			if bv.T != nil {
				rlm.assertTypeIsPublic(store, bv.T, visited)
			}
		}
		if v.Blank.T != nil {
			rlm.assertTypeIsPublic(store, v.Blank.T, visited)
		}
	case *PackageValue:
		if v.PkgPath != rlm.Path && isPkgPrivateFromPkgPath(store, v.PkgPath) {
			panic("cannot persist package from the private realm " + v.PkgPath)
		}
	default:
		panic(fmt.Sprintf("assertNoPrivateDeps: unhandled object type %T", v))
	}
}

// assertTypeIsPublic ensure that the type t is not defined in a private realm.
// it do it recursively for all types in t and have recursive guard to avoid infinite recursion on declared types.
//
// XXX JAE: In addition to the other comment above about assertObjectIsPublic() usage,
// shouldn't this be computed once for every type statically in the preprocessor?
// The type itself could have a boolean, and every type could have SetIsPrivate() (no args)
// after construction that sets the boolean based on its dependencies.
// This slow implementation seems fine for now, something to optimize later.
func (rlm *Realm) assertTypeIsPublic(store Store, t Type, visited map[TypeID]struct{}) {
	pkgPath := ""

	// NOTE: Use to avoid infinite recursion on declared types & avoid repeated checks.
	tid := t.TypeID()
	if _, exists := visited[tid]; exists {
		return
	}
	visited[tid] = struct{}{}
	switch tt := t.(type) {
	case *FuncType:
		for _, param := range tt.Params {
			rlm.assertTypeIsPublic(store, param, visited)
		}
		for _, result := range tt.Results {
			rlm.assertTypeIsPublic(store, result, visited)
		}
	case FieldType:
		rlm.assertTypeIsPublic(store, tt.Type, visited)
	case *SliceType, *ArrayType, *ChanType, *PointerType:
		rlm.assertTypeIsPublic(store, tt.Elem(), visited)
	case *tupleType:
		for _, et := range tt.Elts {
			rlm.assertTypeIsPublic(store, et, visited)
		}
	case *MapType:
		rlm.assertTypeIsPublic(store, tt.Key, visited)
		rlm.assertTypeIsPublic(store, tt.Elem(), visited)
	case *InterfaceType:
		pkgPath = tt.GetPkgPath()
		for _, method := range tt.Methods {
			rlm.assertTypeIsPublic(store, method, visited)
		}
	case *StructType:
		pkgPath = tt.GetPkgPath()
		for _, field := range tt.Fields {
			rlm.assertTypeIsPublic(store, field, visited)
		}
	case *DeclaredType:
		rlm.assertTypeIsPublic(store, tt.Base, visited)
		for _, method := range tt.Methods {
			rlm.assertTypeIsPublic(store, method.T, visited)
			if mv, ok := method.V.(*FuncValue); ok {
				rlm.assertTypeIsPublic(store, mv.Type, visited)
			}
		}
		pkgPath = tt.GetPkgPath()
	case *RefType:
		panic("should not happen: ref type in assert type is public")
	case PrimitiveType, *TypeType, *PackageType, blockType, heapItemType:
		// these types do not have a package path.
		// NOTE: PackageType have a TypeID, should i loat it from store and check it?
		return
	default:
		panic(fmt.Sprintf("assertTypeIsPublic: unhandled type %T", tt))
	}
	if pkgPath != "" && pkgPath != rlm.Path && isPkgPrivateFromPkgPath(store, pkgPath) {
		panic("cannot persist object of type defined in the private realm " + pkgPath)
	}
}

//----------------------------------------
// getSelfOrChildObjects

// Get self (if object) or child objects.
// Value is either Object or RefValue.
// Shallow; doesn't recurse into objects.
func getSelfOrChildObjects(val Value, more []Value) []Value {
	if _, ok := val.(RefValue); ok {
		return append(more, val)
	} else if _, ok := val.(Object); ok {
		return append(more, val)
	} else {
		return getChildObjects(val, more)
	}
}

// Gets child objects.
// Shallow; doesn't recurse into objects.
func getChildObjects(val Value, more []Value) []Value {
	switch cv := val.(type) {
	case nil:
		return more
	case StringValue:
		return more
	case BigintValue:
		return more
	case BigdecValue:
		return more
	case DataByteValue:
		panic("cannot get children from data byte objects")
	case PointerValue:
		if cv.Base == nil {
			panic("should not happen")
		}
		more = getSelfOrChildObjects(cv.Base, more)
		return more
	case *ArrayValue:
		for _, ctv := range cv.List {
			more = getSelfOrChildObjects(ctv.V, more)
		}
		return more
	case *SliceValue:
		more = getSelfOrChildObjects(cv.Base, more)
		return more
	case *StructValue:
		for _, ctv := range cv.Fields {
			more = getSelfOrChildObjects(ctv.V, more)
		}
		return more
	case *FuncValue:
		if bv, ok := cv.Parent.(*Block); ok {
			more = getSelfOrChildObjects(bv, more)
		}
		for _, c := range cv.Captures {
			more = getSelfOrChildObjects(c.V, more)
		}
		return more
	case *BoundMethodValue:
		more = getSelfOrChildObjects(cv.Func, more)
		more = getSelfOrChildObjects(cv.Receiver.V, more)
		return more
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			more = getSelfOrChildObjects(cur.Key.V, more)
			more = getSelfOrChildObjects(cur.Value.V, more)
		}
		return more
	case TypeValue:
		return more
	case *PackageValue:
		more = getSelfOrChildObjects(cv.Block, more)
		for _, fb := range cv.FBlocks {
			more = getSelfOrChildObjects(fb, more)
		}
		return more
	case *Block:
		for _, ctv := range cv.Values {
			more = getSelfOrChildObjects(ctv.V, more)
		}
		// Generally the parent block must also be persisted.
		// Otherwise NamePath may not resolve when referencing
		// a parent block.
		more = getSelfOrChildObjects(cv.Parent, more)
		return more
	case *HeapItemValue:
		more = getSelfOrChildObjects(cv.Value.V, more)
		return more
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

// like getChildObjects() but loads RefValues into objects.
func getChildObjects2(store Store, val Value) []Object {
	chos := getChildObjects(val, nil)
	objs := make([]Object, 0, len(chos))
	for _, child := range chos {
		if ref, ok := child.(RefValue); ok {
			oo := store.GetObject(ref.ObjectID)
			objs = append(objs, oo)
		} else if oo, ok := child.(Object); ok {
			objs = append(objs, oo)
		}
	}
	return objs
}

//----------------------------------------
// getUnsavedChildObjects

// Gets all unsaved child objects.
// Shallow; doesn't recurse into objects.
func getUnsavedChildObjects(val Value) []Object {
	vals := getChildObjects(val, nil)
	unsaved := make([]Object, 0, len(vals))
	for _, val := range vals {
		// sanity check:
		if pv, ok := val.(*PackageValue); ok {
			if !pv.IsRealm() && pv.GetIsDirty() {
				panic("unexpected dirty non-realm package " + pv.PkgPath)
			}
		}
		// ...
		if _, ok := val.(RefValue); ok {
			// is already saved.
		} else if obj, ok := val.(Object); ok {
			// if object...
			if isUnsaved(obj) {
				unsaved = append(unsaved, obj)
			}
		} else {
			panic("unsaved child is not an object")
		}
	}
	return unsaved
}

//----------------------------------------
// copyTypeWithRefs

func copyMethods(methods []TypedValue) []TypedValue {
	res := make([]TypedValue, len(methods))
	for i, mtv := range methods {
		// NOTE: this works because copyMethods/copyTypeWithRefs
		// gets called AFTER the file block (method closure)
		// gets saved (e.g. from *Machine.savePackage()).
		res[i] = TypedValue{
			T: copyTypeWithRefs(mtv.T),
			V: copyValueWithRefs(mtv.V),
		}
	}
	return res
}

func refOrCopyType(typ Type) Type {
	if dt, ok := typ.(*DeclaredType); ok {
		return RefType{ID: dt.TypeID()}
	} else {
		return copyTypeWithRefs(typ)
	}
}

func copyFieldsWithRefs(fields []FieldType) []FieldType {
	fieldsCpy := make([]FieldType, len(fields))
	for i, field := range fields {
		fieldsCpy[i] = FieldType{
			Name:     field.Name,
			Type:     refOrCopyType(field.Type),
			Embedded: field.Embedded,
			Tag:      field.Tag,
		}
	}
	return fieldsCpy
}

// Copies type but with references to dependant types;
// the result is suitable for persistence bytes serialization.
func copyTypeWithRefs(typ Type) Type {
	switch ct := typ.(type) {
	case nil:
		panic("cannot copy nil types")
	case PrimitiveType:
		return ct
	case *PointerType:
		return &PointerType{
			Elt: refOrCopyType(ct.Elt),
		}
	case FieldType:
		panic("cannot copy field types")
	case *ArrayType:
		return &ArrayType{
			Len: ct.Len,
			Elt: refOrCopyType(ct.Elt),
			Vrd: ct.Vrd,
		}
	case *SliceType:
		return &SliceType{
			Elt: refOrCopyType(ct.Elt),
			Vrd: ct.Vrd,
		}
	case *StructType:
		return &StructType{
			PkgPath: ct.PkgPath,
			Fields:  copyFieldsWithRefs(ct.Fields),
		}
	case *FuncType:
		return &FuncType{
			Params:  copyFieldsWithRefs(ct.Params),
			Results: copyFieldsWithRefs(ct.Results),
		}
	case *MapType:
		return &MapType{
			Key:   refOrCopyType(ct.Key),
			Value: refOrCopyType(ct.Value),
		}
	case *InterfaceType:
		return &InterfaceType{
			PkgPath:    ct.PkgPath,
			FieldTypes: copyFieldsWithRefs(ct.FieldTypes),
			Generic:    ct.Generic,
		}
	case *TypeType:
		return &TypeType{}
	case *DeclaredType:
		if ct.PkgPath == uversePkgPath {
			// This happens with a type alias to a uverse type.
			return RefType{ID: ct.TypeID()}
		}
		dt := &DeclaredType{
			PkgPath: ct.PkgPath,
			Name:    ct.Name,
			Base:    copyTypeWithRefs(ct.Base),
			Methods: copyMethods(ct.Methods),
		}
		return dt
	case *PackageType:
		return &PackageType{}
	case *ChanType:
		return &ChanType{
			Dir: ct.Dir,
			Elt: refOrCopyType(ct.Elt),
		}
	case blockType:
		return blockType{}
	case *tupleType:
		elts2 := make([]Type, len(ct.Elts))
		for i, elt := range ct.Elts {
			elts2[i] = refOrCopyType(elt)
		}
		return &tupleType{
			Elts: elts2,
		}
	case RefType:
		return RefType{
			ID: ct.ID,
		}
	case heapItemType:
		return ct
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", typ))
	}
}

//----------------------------------------
// copyValueWithRefs

// Copies value but with references to objects; the result is suitable for
// persistence bytes serialization.
// Also checks for integrity of immediate children -- they must already be
// persistent (real), and not dirty, or else this function panics.
func copyValueWithRefs(val Value) Value {
	switch cv := val.(type) {
	case nil:
		return nil
	case StringValue:
		return cv
	case BigintValue:
		return cv
	case BigdecValue:
		return cv
	case DataByteValue:
		panic("cannot copy data byte value with references")
	case PointerValue:
		if cv.Base == nil {
			panic("should not happen")
		}
		return PointerValue{
			/*
				already represented in .Base[Index]:
				TypedValue: &TypedValue{
					T: cv.TypedValue.T,
					V: copyValueWithRefs(cv.TypedValue.V),
				},
			*/
			Base:  toRefValue(cv.Base),
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = refOrCopyValue(etv)
			}
			return &ArrayValue{
				ObjectInfo: cv.ObjectInfo.Copy(),
				List:       list,
			}
		} else {
			return &ArrayValue{
				ObjectInfo: cv.ObjectInfo.Copy(),
				Data:       cp(cv.Data),
			}
		}
	case *SliceValue:
		return &SliceValue{
			Base:   toRefValue(cv.Base),
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = refOrCopyValue(ftv)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			parent = toRefValue(cv.Parent)
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = refOrCopyValue(ctv)
		}
		// nativeBody funcs which don't come from NativeResolver (and thus don't
		// have NativePkg/Name) can't be persisted, and should not be able
		// to get here anyway.
		if cv.nativeBody != nil && cv.NativePkg == "" {
			panic("cannot copy function value with native body when there is no native package")
		}
		ft := copyTypeWithRefs(cv.Type)
		return &FuncValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Type:       ft,
			IsMethod:   cv.IsMethod,
			Source:     source,
			Name:       cv.Name,
			Parent:     parent,
			Captures:   captures,
			FileName:   cv.FileName,
			PkgPath:    cv.PkgPath,
			NativePkg:  cv.NativePkg,
			NativeName: cv.NativeName,
			Crossing:   cv.Crossing,
		}
	case *BoundMethodValue:
		fnc := copyValueWithRefs(cv.Func).(*FuncValue)
		rtv := refOrCopyValue(cv.Receiver)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := refOrCopyValue(cur.Key)
			val2 := refOrCopyValue(cur.Value)
			list.Append(nilAllocator, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		return toTypeValue(copyTypeWithRefs(cv.Type))
	case *PackageValue:
		block := toRefValue(cv.Block)
		fblocks := make([]Value, len(cv.FBlocks))
		for i, fb := range cv.FBlocks {
			fblocks[i] = toRefValue(fb)
		}
		return &PackageValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Block:      block,
			PkgName:    cv.PkgName,
			PkgPath:    cv.PkgPath,
			Private:    cv.Private,
			FNames:     cv.FNames, // no copy
			FBlocks:    fblocks,
			Realm:      cv.Realm,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = refOrCopyValue(tv)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = toRefValue(cv.Parent)
		}
		bl := &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{}, // empty
		}
		return bl
	case RefValue:
		return cv
	case *HeapItemValue:
		hiv := &HeapItemValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Value:      refOrCopyValue(cv.Value),
		}
		return hiv
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

//----------------------------------------
// fillTypes

// (fully) fills the type.
// The store stores RefTypes, but this function fills it.
// This lets the store be independent of laziness.
func fillType(store Store, typ Type) Type {
	switch ct := typ.(type) {
	case nil:
		return nil
	case PrimitiveType:
		return ct
	case *PointerType:
		ct.Elt = fillType(store, ct.Elt)
		return ct
	case FieldType:
		panic("cannot fill field types")
	case *ArrayType:
		ct.Elt = fillType(store, ct.Elt)
		return ct
	case *SliceType:
		ct.Elt = fillType(store, ct.Elt)
		return ct
	case *StructType:
		for i, field := range ct.Fields {
			ct.Fields[i].Type = fillType(store, field.Type)
		}
		return ct
	case *FuncType:
		for i, param := range ct.Params {
			ct.Params[i].Type = fillType(store, param.Type)
		}
		for i, result := range ct.Results {
			ct.Results[i].Type = fillType(store, result.Type)
		}
		return ct
	case *MapType:
		ct.Key = fillType(store, ct.Key)
		ct.Value = fillType(store, ct.Value)
		return ct
	case *InterfaceType:
		for i, mthd := range ct.FieldTypes {
			ct.FieldTypes[i].Type = fillType(store, mthd.Type)
		}
		return ct
	case *TypeType:
		return ct // nothing to do
	case *DeclaredType:
		// replace ct with store type.
		tid := ct.TypeID()
		ct = nil // defensive.
		ct2 := store.GetType(tid).(*DeclaredType)
		if ct2.sealed {
			// recursion protection needed for methods that reference
			// the declared type recursively (and ct != ct2).
			// use the sealed flag for recursion protection.
			return ct2
		} else {
			ct2.sealed = true
			ct2.Base = fillType(store, ct2.Base)
			for i, method := range ct2.Methods {
				ct2.Methods[i].T = fillType(store, method.T)
				mv := ct2.Methods[i].V.(*FuncValue)
				mv.Type = fillType(store, mv.Type)
			}
			return ct2
		}
	case *PackageType:
		return ct // nothing to do
	case *ChanType:
		ct.Elt = fillType(store, ct.Elt)
		return ct
	case blockType:
		return ct // nothing to do
	case *tupleType:
		for i, elt := range ct.Elts {
			ct.Elts[i] = fillType(store, elt)
		}
		return ct
	case RefType:
		return store.GetType(ct.TypeID())
	case heapItemType:
		return ct
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", reflect.TypeOf(typ)))
	}
}

func fillTypesTV(store Store, tv *TypedValue) {
	tv.T = fillType(store, tv.T)
	tv.V = fillTypesOfValue(store, tv.V)
}

// Partially fills loaded objects shallowly, similarly to
// getUnsavedTypes. Replaces all RefTypes with corresponding types.
func fillTypesOfValue(store Store, val Value) Value {
	switch cv := val.(type) {
	case nil: // do nothing
		return cv
	case StringValue: // do nothing
		return cv
	case BigintValue: // do nothing
		return cv
	case BigdecValue: // do nothing
		return cv
	case DataByteValue: // do nothing
		return cv
	case PointerValue:
		if cv.Base != nil {
			// cv.Base is object.
			// fillTypesOfValue(store, cv.Base) (wrong)
			return cv
		} else {
			fillTypesTV(store, cv.TV)
			return cv
		}
	case *ArrayValue:
		for i := range cv.List {
			ctv := &cv.List[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *SliceValue:
		fillTypesOfValue(store, cv.Base)
		return cv
	case *StructValue:
		for i := range cv.Fields {
			ctv := &cv.Fields[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *FuncValue:
		cv.Type = fillType(store, cv.Type)
		return cv
	case *BoundMethodValue:
		fillTypesOfValue(store, cv.Func)
		fillTypesTV(store, &cv.Receiver)
		return cv
	case *MapValue:
		cv.vmap = make(map[MapKey]*MapListItem, cv.List.Size)
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			fillTypesTV(store, &cur.Key)
			fillTypesTV(store, &cur.Value)

			fillValueTV(store, &cur.Key)
			mk, isNaN := cur.Key.ComputeMapKey(store, false)
			if !isNaN {
				cv.vmap[mk] = cur
			}
		}
		return cv
	case TypeValue:
		cv.Type = fillType(store, cv.Type)
		return cv
	case *PackageValue:
		fillTypesOfValue(store, cv.Block)
		return cv
	case *Block:
		for i := range cv.Values {
			ctv := &cv.Values[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case RefValue: // do nothing
		return cv
	case *HeapItemValue:
		fillTypesTV(store, &cv.Value)
		return cv
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

//----------------------------------------
// persistence

func (rlm *Realm) nextObjectID() ObjectID {
	if rlm == nil {
		panic("cannot get next object ID of nil realm")
	}
	if rlm.ID.IsZero() {
		panic("cannot get next object ID of realm without an ID")
	}
	rlm.Time++
	nxtid := ObjectID{
		PkgID:   rlm.ID,
		NewTime: rlm.Time, // starts at 1.
	}
	return nxtid
}

// Object gets its id set (panics if already set), and becomes
// marked as new and real.
func (rlm *Realm) assignNewObjectID(oo Object) ObjectID {
	oid := oo.GetObjectID()
	if !oid.IsZero() {
		panic("unexpected non-zero object id")
	}
	noid := rlm.nextObjectID()
	oo.SetObjectID(noid)
	return noid
}

//----------------------------------------
// Misc.

// should not be used outside of realm.go
func toRefNode(bn BlockNode) RefNode {
	return RefNode{
		Location:  bn.GetLocation(),
		BlockNode: nil, // NOTE is always nil.
	}
}

// should not be used outside of realm.go
func toRefValue(val Value) RefValue {
	// TODO use type switch stmt.
	if ref, ok := val.(RefValue); ok {
		return ref
	} else if oo, ok := val.(Object); ok {
		if pv, ok := val.(*PackageValue); ok {
			if pv.GetIsDirty() {
				panic("unexpected dirty package " + pv.PkgPath)
			}
			return RefValue{
				PkgPath: pv.PkgPath,
			}
		} else if !oo.GetIsReal() {
			panic("unexpected unreal object")
		}
		// This can happen with some circular
		// references.
		// else if oo.GetIsDirty() {
		// panic("unexpected dirty object")
		// }
		if oo.GetIsNewEscaped() {
			// NOTE: oo.GetOwnerID() will become zero.
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Escaped:  true,
				// Hash: nil,
			}
		} else if oo.GetIsEscaped() {
			if debugRealm {
				if !oo.GetOwnerID().IsZero() {
					panic("cannot convert escaped object to ref value without an owner ID")
				}
			}
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Escaped:  true,
				// Hash: nil,
			}
		} else {
			if debugRealm {
				if oo.GetRefCount() > 1 {
					panic("unexpected references when converting to ref value")
				}
				if oo.GetHash().IsZero() {
					panic("hash missing when converting to ref value")
				}
			}
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Hash:     oo.GetHash(),
			}
		}
	} else {
		panic("unexpected error converting to ref value")
	}
}

func ensureUniq(oozz ...[]Object) {
	count := 0
	for _, ooz := range oozz {
		count += len(ooz)
	}
	om := make(map[Object]struct{}, count) // TODO: count*constant?
	for _, ooz := range oozz {
		for _, uo := range ooz {
			if _, ok := om[uo]; ok {
				panic(fmt.Sprintf("duplicate object: %v", uo))
			} else {
				om[uo] = struct{}{}
			}
		}
	}
}

func refOrCopyValue(tv TypedValue) TypedValue {
	if tv.T != nil {
		tv.T = refOrCopyType(tv.T)
	}
	if obj, ok := tv.V.(Object); ok {
		tv.V = toRefValue(obj)
		return tv
	} else {
		tv.V = copyValueWithRefs(tv.V)
		return tv
	}
}

func isUnsaved(oo Object) bool {
	return oo.GetIsNewReal() || oo.GetIsDirty()
}

func prettyJSON(jstr []byte) []byte {
	var c any
	err := json.Unmarshal(jstr, &c)
	if err != nil {
		return nil
	}
	js, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return nil
	}
	return js
}

func getOwner(store Store, oo Object) Object {
	po := oo.GetOwner()
	poid := oo.GetOwnerID()
	if po == nil {
		if !poid.IsZero() {
			po = store.GetObject(poid)
			oo.SetOwner(po)
		}
	}
	return po
}

// XXX this would be a lot faster if the PkgID itself included a private bit;
// no store argument or lookup would be needed.
func isPkgPrivateFromPkgID(store Store, pkgID PkgID) bool {
	oid := ObjectIDFromPkgID(pkgID)
	oo := store.GetObject(oid)
	pv, ok := oo.(*PackageValue)
	if !ok {
		panic("oid with time set at 1 does not refer to package value")
	}
	return pv.Private
}

func pkgPathFromPkgID(store Store, pkgID PkgID) string {
	oid := ObjectIDFromPkgID(pkgID)
	oo := store.GetObject(oid)
	pv, ok := oo.(*PackageValue)
	if !ok {
		panic("oid with time set at 1 does not refer to package value")
	}
	return pv.PkgPath
}

func isPkgPrivateFromPkgPath(store Store, pkgPath string) bool {
	pv := store.GetPackage(pkgPath, false)
	if pv == nil {
		panic("cannot find package value from store for path " + pkgPath)
	}
	return pv.Private
}
