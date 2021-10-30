package gno

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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

Every "realm package" should define at last one package-level variable:

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
// Realm

type RealmID struct {
	Hashlet
}

func (rid RealmID) MarshalAmino() (string, error) {
	return hex.EncodeToString(rid.Hashlet[:]), nil
}

func (rid *RealmID) UnmarshalAmino(h string) error {
	_, err := hex.Decode(rid.Hashlet[:], []byte(h))
	return err
}

func (rid RealmID) String() string {
	return fmt.Sprintf("RID%X", rid.Hashlet[:])
}

func (rid RealmID) Bytes() []byte {
	return rid.Hashlet[:]
}

func RealmIDFromPath(path string) RealmID {
	return RealmID{HashBytes([]byte(path))}
}

func ObjectIDFromPkgPath(path string) ObjectID {
	return ObjectID{
		RealmID: RealmIDFromPath(path),
		NewTime: 1, // by realm logic.
	}
}

// NOTE: A nil realm is special and has limited functionality; enough to
// support methods that don't require persistence. This is the default realm
// when a machine starts with a non-realm package.
type Realm struct {
	ID   RealmID
	Path string
	Time uint64

	newCreated []Object
	newEscaped []Object
	newDeleted []Object

	created []Object // about to become real.
	updated []Object // real objects that were modified.
	deleted []Object // real objects that became deleted.
	escaped []Object // real objects with refcount > 1.
}

// Creates a blank new realm with counter 0.
func NewRealm(path string) *Realm {
	id := RealmIDFromPath(path)
	return &Realm{
		ID:   id,
		Path: path,
		Time: 0,
	}
}

func (rlm *Realm) String() string {
	if rlm == nil {
		return "Realm(nil)"
	} else {
		return fmt.Sprintf(
			"Realm{Path:%q:Time:%d}#%X",
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
	if rlm == nil {
		return
	}
	if debug {
		if co != nil && co.GetIsDeleted() {
			panic("cannot attach a deleted object")
		}
		if po != nil && po.GetIsTransient() {
			panic("should not happen")
		}
		if po != nil && po.GetIsDeleted() {
			panic("cannot attach to a deleted object")
		}
	}
	if po == nil || !po.GetIsReal() {
		return // do nothing.
	}
	// From here on, po is real (not new-real).
	// Updates to .newCreated/.newEscaped /.newDeleted made here. (first gen)
	// More appends happen during FinalizeRealmTransactions(). (second+ gen)
	rlm.MarkDirty(po)
	if co != nil {
		co.IncRefCount()
		if co.GetRefCount() > 1 {
			if co.GetIsEscaped() {
				// already escaped
			} else {
				rlm.MarkNewEscaped(co)
			}
		} else if co.GetIsReal() {
			rlm.MarkDirty(co)
		} else {
			co.SetOwner(po)
			rlm.MarkNewReal(co)
		}
	}
	if xo != nil {
		xo.DecRefCount()
		if xo.GetRefCount() == 0 {
			if xo.GetIsNewReal() || xo.GetIsReal() {
				rlm.MarkNewDeleted(xo)
			}
		}
	}
}

//----------------------------------------
// mark*

func (rlm *Realm) MarkNewReal(oo Object) {
	if debug {
		if pv, ok := oo.(*PackageValue); ok {
			// packages should have no owner.
			if pv.GetOwner() != nil {
				panic("should not happen")
			}
			// packages should have ref-count 1.
			if pv.GetRefCount() != 1 {
				panic("should not happen")
			}
		} else {
			if oo.GetOwner() == nil {
				panic("should not happen")
			}
			if !oo.GetOwner().GetIsReal() {
				panic("should not happen")
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
	if debug {
		if !oo.GetIsReal() && !oo.GetIsNewReal() {
			panic("should not happen")
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
	if debug {
		if !oo.GetIsNewReal() && !oo.GetIsReal() {
			panic("should not happen")
		}
		if oo.GetIsDeleted() {
			panic("should not happen")
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
	if debug {
		if !oo.GetIsNewReal() && !oo.GetIsReal() {
			panic("should not happen")
		}
		if oo.GetIsDeleted() {
			panic("should not happen")
		}
		if oo.GetIsEscaped() {
			panic("should not happen")
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
func (rlm *Realm) FinalizeRealmTransaction(readonly bool, store Store) {
	if readonly {
		if true ||
			len(rlm.newCreated) > 0 ||
			len(rlm.newEscaped) > 0 ||
			len(rlm.newDeleted) > 0 ||
			len(rlm.created) > 0 ||
			len(rlm.updated) > 0 ||
			len(rlm.deleted) > 0 ||
			len(rlm.escaped) > 0 {
			panic("realm updates in readonly transaction")
		}
		return
	}
	if debug {
		// * newCreated - may become created unless ancestor is deleted
		// * newDeleted - may become deleted unless attached to new-real owner
		// * newEscaped - may become escaped unless new-real and refcount 0 or 1.
		// * updated - includes all real updated objects, and will be appended with ancestors
		ensureUniq(rlm.newCreated)
		ensureUniq(rlm.newEscaped)
		ensureUniq(rlm.newDeleted)
		ensureUniq(rlm.updated)
		if false ||
			rlm.created != nil ||
			rlm.deleted != nil ||
			rlm.escaped != nil {
			panic("should not happen")
		}
	}

	// increment recursively for created descendants.
	// also assigns object ids for all.
	rlm.processNewCreatedMarks(store)
	// decrement recursively for deleted descendants.
	rlm.processNewDeletedMarks(store)
	// at this point, all ref-counts are final.
	// demote any escaped if ref-count is 1.
	rlm.processNewEscapedMarks(store)
	// given created and updated objects,
	// mark all owned-ancestors also as dirty.
	rlm.markDirtyAncestors(store)
	if debug {
		fmt.Println("newCreated", rlm.newCreated)
		fmt.Println("newEscaped", rlm.newEscaped)
		fmt.Println("newDeleted", rlm.newDeleted)
		fmt.Println("created", rlm.created)
		fmt.Println("updated", rlm.updated)
		fmt.Println("deleted", rlm.deleted)
		fmt.Println("escaped", rlm.escaped)
		ensureUniq(rlm.created, rlm.updated, rlm.deleted)
		ensureUniq(rlm.escaped)
	}
	// save all the created and updated objects.
	// hash calculation is done along the way,
	// or via escaped-object persistence in
	// the iavl tree.
	rlm.saveUnsavedObjects(store)
	// delete all deleted objects.
	rlm.removeDeletedObjects(store)
	// reset realm state for new transaction.
	rlm.clearMarks()
}

//----------------------------------------
// processNewCreatedMarks

// Crawls marked created children and increments ref counts,
// finding more newly created objects recursively.
// All newly created objects become appended to .created,
// and get assigned ids.
func (rlm *Realm) processNewCreatedMarks(store Store) {
	for _, oo := range rlm.newCreated {
		if debug {
			if oo.GetIsDirty() {
				panic("should not happen")
			}
		}
		if oo.GetRefCount() == 0 {
			if debug {
				if !oo.GetIsNewDeleted() {
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
}

// oo must be marked new-real, and ref-count already incremented.
func (rlm *Realm) incRefCreatedDescendants(store Store, oo Object) {
	if debug {
		if oo.GetIsDirty() {
			panic("should not happen")
		}
		if oo.GetRefCount() <= 0 {
			panic("should not happen")
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
			if debug {
				if child.GetRefCount() < 1 {
					panic("should not happen")
				}
			}
			// package values are skipped.
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
				rlm.incRefCreatedDescendants(store, child)
				child.SetIsNewReal(true)
			}
		} else if rc > 1 {
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
			panic("should not happen")
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
	if debug {
		if oo.GetObjectID().IsZero() {
			panic("should not happen")
		}
		if oo.GetRefCount() != 0 {
			panic("should not happen")
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
			// do nothing
		} else {
			panic("should not happen")
		}
	}
}

//----------------------------------------
// processNewEscapedMarks

// demotes new-real escaped objects with refcount 0 or 1.  remaining
// objects get their original owners marked dirty (to be further
// marked via markDirtyAncestors).
func (rlm *Realm) processNewEscapedMarks(store Store) {
	escaped := make([]Object, 0, len(rlm.newEscaped))
	// These are those marked by MarkNewEscaped(),
	// regardless of whether new-real or was real,
	// but is always newly escaped,
	// (and never can be unescaped,)
	// except for new-reals that get demoted
	// because ref-count isn't >= 2.
	for _, eo := range rlm.newEscaped {
		if debug {
			if !eo.GetIsNewEscaped() {
				panic("should not happen")
			}
			if eo.GetIsEscaped() {
				panic("should not happen")
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
					panic("should not happen")
				}
				// escaped has no owner.
				eo.SetOwner(nil)
			}
		}
	}
	rlm.escaped = escaped // XXX is this actually used?
}

//----------------------------------------
// markDirtyAncestors

// New and modified objects' owners and their owners
// (ancestors) must be marked as dirty to update the
// hash tree.
func (rlm *Realm) markDirtyAncestors(store Store) {
	markAncestorsOne := func(oo Object) {
		for {
			if pv, ok := oo.(*PackageValue); ok {
				if debug {
					if pv.GetRefCount() < 1 {
						panic("expected package value to have refcount 1 or greater")
					}
				}
				// package values have no ancestors.
				break
			}
			rc := oo.GetRefCount()
			if debug {
				if rc == 0 {
					panic("should not happen")
				}
			}
			if rc > 1 {
				if debug {
					if !oo.GetIsEscaped() && !oo.GetIsNewEscaped() {
						panic("should not happen")
					}
					if !oo.GetOwnerID().IsZero() {
						panic("should not happen")
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
				// via call to markAncestorsOne
				// via .created.
				break
			} else if po.GetIsDirty() {
				// already will be marked
				// via call to markAncestorsOne
				// via .updated.
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
		markAncestorsOne(oo)
	}
	// NOTE: must happen after iterating over rlm.updated
	// for the same reason.
	for _, oo := range rlm.created {
		markAncestorsOne(oo)
	}
}

//----------------------------------------
// saveUnsavedObjects

// Saves .created and .updated objects.
func (rlm *Realm) saveUnsavedObjects(store Store) {
	for _, co := range rlm.created {
		// for i := len(rlm.created) - 1; i >= 0; i-- {
		// co := rlm.created[i]
		if !co.GetIsNewReal() {
			// might have happened already as child
			// of something else created.
			continue
		} else {
			rlm.saveUnsavedObjectRecursively(store, co)
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
			rlm.saveUnsavedObjectRecursively(store, uo)
		}
	}
}

// store unsaved children first.
func (rlm *Realm) saveUnsavedObjectRecursively(store Store, oo Object) {
	if debug {
		if !oo.GetIsNewReal() && !oo.GetIsDirty() {
			panic("should not happen")
		}
		// object id should have been assigned during processNewCreatedMarks.
		if oo.GetObjectID().IsZero() {
			panic("should not happen")
		}
		// deleted objects should not have gotten here.
		if false ||
			oo.GetRefCount() <= 0 ||
			oo.GetIsNewDeleted() ||
			oo.GetIsDeleted() {
			panic("should not happen")
		}
	}
	// first, save unsaved children.
	unsaved := getUnsavedChildObjects(oo)
	for _, uch := range unsaved {
		if uch.GetIsEscaped() || uch.GetIsNewEscaped() {
			// no need to save preemptively.
		} else {
			rlm.saveUnsavedObjectRecursively(store, uch)
		}
	}
	// then, save self.
	if oo.GetIsNewReal() {
		// save created object.
		if debug {
			if oo.GetIsDirty() {
				panic("should not happen")
			}
		}
		rlm.saveObject(store, oo)
		oo.SetIsNewReal(false)
	} else {
		// update existing object.
		if debug {
			if !oo.GetIsDirty() {
				panic("should not happen")
			}
			if !oo.GetIsReal() {
				panic("should not happen")
			}
			if oo.GetIsNewReal() {
				panic("should not happen")
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
	// scan for any types.
	types := getUnsavedTypes(oo, nil)
	for _, typ := range types {
		store.SetType(typ)
	}
	// set hash to escape index.
	if oo.GetIsNewEscaped() {
		oo.SetIsNewEscaped(false)
		oo.SetIsEscaped(true)
		// XXX anything else to do?
	}
	// set object to store.
	// NOTE: also sets the hash to object.
	store.SetObject(oo)
	// set index.
	if oo.GetIsEscaped() {
		// XXX save iod->hash to iavl.
		fmt.Println("XXX save hash to iavl")
	}
}

//----------------------------------------
// removeDeletedObjects

func (rlm *Realm) removeDeletedObjects(store Store) {
	for _, do := range rlm.deleted {
		store.DelObject(do)
	}
}

//----------------------------------------
// clearMarks

func (rlm *Realm) clearMarks() {
	// sanity check
	if debug {
		for _, oo := range rlm.newDeleted {
			if oo.GetIsNewDeleted() {
				panic("should not happen")
			}
		}
		for _, oo := range rlm.newCreated {
			if oo.GetIsNewReal() {
				panic("should not happen")
			}
		}
		for _, oo := range rlm.newEscaped {
			if oo.GetIsNewEscaped() {
				panic("should not happen")
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
	case DataByteValue:
		panic("should not happen")
	case PointerValue:
		if cv.Base != nil {
			more = getSelfOrChildObjects(cv.Base, more)
		} else {
			more = getSelfOrChildObjects(cv.TV.V, more)
		}
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
		if bv, ok := cv.Closure.(*Block); ok {
			more = getSelfOrChildObjects(bv, more)
		}
		return more
	case *BoundMethodValue:
		more = getChildObjects(cv.Func, more) // *FuncValue not object
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
		more = getSelfOrChildObjects(cv.Parent, more)
		return more
	case *NativeValue:
		panic("native values not supported")
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
			panic("should not happen")
		}
	}
	return unsaved
}

//----------------------------------------
// getUnsavedTypes

func getUnsavedType(tt Type, more []Type) []Type {
	if tt.GetIsSaved() {
		return more
	}
	more = append(more, tt)
	return more
}

func getUnsavedTypesTV(tv TypedValue, more []Type) []Type {
	if tv.T != nil {
		more = getUnsavedType(tv.T, more)
	}
	if _, ok := tv.V.(Object); ok {
		// do not recurse into objects.
	} else {
		more = getUnsavedTypes(tv.V, more)
	}
	return more
}

// Get unsaved types from a value.
// Shallow; doesn't recurse into objects.
func getUnsavedTypes(val Value, more []Type) []Type {
	switch cv := val.(type) {
	case nil:
		return more
	case StringValue:
		return more
	case BigintValue:
		return more
	case DataByteValue:
		panic("should not happen")
	case PointerValue:
		if cv.Base != nil {
			// cv.Base by reference.
			// more = getUnsavedTypes(cv.Base, more) (wrong)
		} else {
			more = getUnsavedTypesTV(*cv.TV, more)
		}
		return more
	case *ArrayValue:
		for _, ctv := range cv.List {
			more = getUnsavedTypesTV(ctv, more)
		}
		return more
	case *SliceValue:
		more = getUnsavedTypes(cv.Base, more)
		return more
	case *StructValue:
		for _, ctv := range cv.Fields {
			more = getUnsavedTypesTV(ctv, more)
		}
		return more
	case *FuncValue:
		more = getUnsavedType(cv.Type, more)
		/* XXX prob wrong, as closure is object:
		if bv, ok := cv.Closure.(*Block); ok {
			more = getUnsavedTypes(bv, more)
		}
		*/
		return more
	case *BoundMethodValue:
		more = getUnsavedTypes(cv.Func, more)
		more = getUnsavedTypesTV(cv.Receiver, more)
		return more
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			more = getUnsavedTypesTV(cur.Key, more)
			more = getUnsavedTypesTV(cur.Value, more)
		}
		return more
	case TypeValue:
		more = getUnsavedType(cv.Type, more)
		return more
	case *PackageValue:
		/* XXX prob wrong, as Block and FBlocks are objects:
		more = getUnsavedTypes(cv.Block, more)
		for _, fb := range cv.FBlocks {
			more = getUnsavedTypes(fb, more)
		}
		*/
		return more
	case *Block:
		for _, ctv := range cv.Values {
			more = getUnsavedTypesTV(ctv, more)
		}
		// XXX prob wrong
		// more = getUnsavedTypes(cv.Parent, more)
		return more
	case *NativeValue:
		panic("native values not supported")
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

//----------------------------------------
// copyWithRefs

// Copies value but with references to objects; the result is suitable for
// persistence bytes serialization.
// Also checks for integrity of immediate children -- they must already be
// persistend (real), and not dirty, or else this function panics.
func copyWithRefs(parent Object, val Value) Value {
	switch cv := val.(type) {
	case nil:
		return nil
	case StringValue:
		return cv
	case BigintValue:
		return cv
	case DataByteValue:
		panic("should not happen")
	case PointerValue:
		if cv.Base != nil {
			return PointerValue{
				/*
					already represented in .Base[Index]:
					TypedValue: &TypedValue{
						T: cv.TypedValue.T,
						V: copyWithRefs(cv.TypedValue.V),
					},
				*/
				Base:  toRefValue(parent, cv.Base),
				Index: cv.Index,
			}
		} else {
			etv := refOrCopy(parent, *cv.TV)
			return PointerValue{
				TV: &etv,
				/*
					Base:  nil,
					Index: 0,
				*/
			}
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = refOrCopy(cv, etv)
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
			Base:   toRefValue(parent, cv.Base),
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = refOrCopy(cv, ftv)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var closure Value
		if cv.Closure != nil {
			closure = toRefValue(parent, cv.Closure)
		}
		if cv.nativeBody != nil {
			panic("should not happen")
		}
		ft := RefType{ID: cv.Type.TypeID()}
		return &FuncValue{
			Type:     ft,
			IsMethod: cv.IsMethod,
			Source:   source,
			Name:     cv.Name,
			Closure:  closure,
			FileName: cv.FileName,
			PkgPath:  cv.PkgPath,
		}
	case *BoundMethodValue:
		fnc := copyWithRefs(cv, cv.Func).(*FuncValue)
		rtv := refOrCopy(cv, cv.Receiver)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(), // XXX ???
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := refOrCopy(cv, cur.Key)
			val2 := refOrCopy(cv, cur.Value)
			list.Append(key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		return toTypeValue(RefType{
			ID: cv.Type.TypeID(),
		})
	case *PackageValue:
		block := toRefValue(cv, cv.Block)
		fblocks := make([]Value, len(cv.FBlocks))
		for i, fb := range cv.FBlocks {
			fblocks[i] = toRefValue(cv, fb)
		}
		return &PackageValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Block:      block,
			PkgName:    cv.PkgName,
			PkgPath:    cv.PkgPath,
			FNames:     cv.FNames, // no copy
			FBlocks:    fblocks,
			Realm:      cv.Realm,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = refOrCopy(cv, tv)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = toRefValue(parent, cv.Parent)
		}
		return &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{}, // empty
		}
	case *NativeValue:
		panic("native values not supported")
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

//----------------------------------------
// fillTypes

func fillTypePtr(store Store, ptr *Type) {
	if *ptr != nil {
		*ptr = store.GetType((*ptr).(RefType).TypeID())
	}
}

func fillTypesTV(store Store, tv *TypedValue) {
	if tvt, ok := tv.T.(RefType); ok {
		tv.T = store.GetType(tvt.TypeID())
	}
	tv.V = fillTypes(store, tv.V)
}

// Partially fills loaded objects shallowly, similarly to
// getUnsavedTypes.  Replaces all RefTypes with corresponding types.
func fillTypes(store Store, val Value) Value {
	switch cv := val.(type) {
	case nil: // do nothing
		return cv
	case StringValue: // do nothing
		return cv
	case BigintValue: // do nothing
		return cv
	case DataByteValue: // do nothing
		return cv
	case PointerValue:
		if cv.Base != nil {
			// cv.Base is object.
			// fillTypes(store, cv.Base) (wrong)
			return cv
		} else {
			fillTypesTV(store, cv.TV)
			return cv
		}
	case *ArrayValue:
		for i := 0; i < len(cv.List); i++ {
			ctv := &cv.List[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *SliceValue:
		fillTypes(store, cv.Base)
		return cv
	case *StructValue:
		for i := 0; i < len(cv.Fields); i++ {
			ctv := &cv.Fields[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *FuncValue:
		fillTypePtr(store, &cv.Type)
		return cv
	case *BoundMethodValue:
		fillTypes(store, cv.Func)
		fillTypesTV(store, &cv.Receiver)
		return cv
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			fillTypesTV(store, &cur.Key)
			fillTypesTV(store, &cur.Value)
		}
		return cv
	case TypeValue:
		fillTypePtr(store, &cv.Type)
		return cv
	case *PackageValue:
		fillTypes(store, cv.Block)
		return cv
	case *Block:
		for i := 0; i < len(cv.Values); i++ {
			ctv := &cv.Values[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *NativeValue:
		panic("native values not supported")
	case RefValue: // do nothing
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
		panic("should not happen")
	}
	if rlm.ID.IsZero() {
		panic("should not happen")
	}
	rlm.Time++
	return ObjectID{
		RealmID: rlm.ID,
		NewTime: rlm.Time, // starts at 1.
	}
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

func toRefNode(bn BlockNode) RefNode {
	return RefNode{
		Location:  bn.GetLocation(),
		BlockNode: nil, // NOTE is always nil.
	}
}

func toRefValue(parent Object, val Value) RefValue {
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
		} else if oo.GetIsDirty() {
			// This can happen with some circular
			// references.
			// panic("unexpected dirty object")
		}
		if oo.GetIsNewEscaped() {
			// NOTE: oo.GetOwnerID() will become zero.
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Escaped:  true,
				// Hash: nil,
			}
		} else if oo.GetIsEscaped() {
			if debug {
				if !oo.GetOwnerID().IsZero() {
					panic("should not happen")
				}
			}
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Escaped:  true,
				// Hash: nil,
			}
		} else {
			if debug {
				if oo.GetRefCount() > 1 {
					panic("should not happen")
				}
				if oo.GetHash().IsZero() {
					panic("should not happen")
				}
			}
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Hash:     oo.GetHash(),
			}
		}
	} else {
		panic("should not happen")
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
				panic("duplicate object")
			} else {
				om[uo] = struct{}{}
			}
		}
	}
}

func refOrCopy(parent Object, tv TypedValue) TypedValue {
	if tv.T != nil {
		tv.T = RefType{tv.T.TypeID()}
	}
	if obj, ok := tv.V.(Object); ok {
		tv.V = toRefValue(parent, obj)
		return tv
	} else {
		tv.V = copyWithRefs(parent, tv.V)
		return tv
	}
}

func isUnsaved(oo Object) bool {
	return oo.GetIsNewReal() || oo.GetIsDirty()
}

func IsRealmPath(pkgPath string) bool {
	// TODO: make it more distinct to distinguish from normal paths.
	if strings.HasPrefix(pkgPath, "gno.land/r/") {
		return true
	} else {
		return false
	}
}

func prettyJSON(jstr []byte) []byte {
	var c interface{}
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
