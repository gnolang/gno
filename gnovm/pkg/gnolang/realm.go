package gnolang

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
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

// NOTE: A nil realm is special and has limited functionality; enough to
// support methods that don't require persistence. This is the default realm
// when a machine starts with a non-realm package.
type Realm struct {
	ID   PkgID
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
	id := PkgIDFromPkgPath(path)
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
			"Realm{Path:%q,Time:%d}#%X",
			rlm.Path, rlm.Time, rlm.ID.Bytes())
	}
}

//----------------------------------------
// ownership hooks

// ref value is the derived value from co, like a slice.
func (rlm *Realm) DidUpdate(store Store, po, xo, co Object) {
	debug2.Printf2("DidUpdate, po: %v, type of po: %v\n", po, reflect.TypeOf(po))
	debug2.Println2("po.GetIsReal: ", po.GetIsReal())
	debug2.Printf2("xo: %v, type of xo: %v\n", xo, reflect.TypeOf(xo))
	debug2.Printf2("co: %v, type of co: %v\n", co, reflect.TypeOf(co))
	//debug2.Println2("rlm.ID: ", rlm.ID)
	if co != nil {
		debug2.Println2("co.GetOriginRealm: ", co.GetOriginRealm())
		debug2.Println2("co.GetIsRef: ", co.GetIsRef())
		debug2.Println2("co.GetRefCount: ", co.GetRefCount())
		debug2.Println2("co.GetIsReal: ", co.GetIsReal())
	}

	if rlm == nil {
		return
	}
	if debug {
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

	if po == nil || !po.GetIsReal() { // XXX, make sure po is attached
		debug2.Println2("po not real, do nothing!!!")
		return // do nothing.
	}

	// TODO: check unreal external here, if po is real, association is invalid, panic

	// else, defer to finalize???
	debug2.Println2("po.GetObjectID().PkgID: ", po.GetObjectID().PkgID)
	if po.GetObjectID().PkgID != rlm.ID {
		panic("cannot modify external-realm or non-realm object")
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
		// XXX, inc ref count everytime assignment happens
		co.IncRefCount()
		if co.GetRefCount() > 1 {
			rlm.MarkNewEscapedCheckCrossRealm(store, co)
		} else {
			if co.GetIsReal() { // TODO: how this happen?
				rlm.MarkDirty(co)
			} else {
				debug2.Println2("set owner of co: ", co)
				co.SetOwner(po)
				rlm.MarkNewReal(co)
			}
			// check cross realm for non escaped objects
			debug2.Println2("=========po is real, check cross realm for non escaped objects========")
			checkCrossRealm(rlm, store, co, false)
		}
	}

	if xo != nil {
		xo.DecRefCount()
		if xo.GetRefCount() == 0 {
			if xo.GetIsReal() {
				rlm.MarkNewDeleted(xo)
			}
		}
	}
}

// check cross realm recursively
func checkCrossRealm2(rlm *Realm, store Store, tv *TypedValue, isLastRef bool) {
	debug2.Println2("checkCrossRealm2, tv: ", tv, reflect.TypeOf(tv.V))
	tv2 := fillValueTV(store, tv)
	if oo, ok := tv2.V.(Object); ok {
		debug2.Println2("is object")
		// set origin realm for embedded value
		oo.SetOriginRealm(tv2.GetOriginPkg(store))
		checkCrossRealm(rlm, store, oo, isLastRef)
	} else { // reference to object
		switch tv.V.(type) {
		case *SliceValue, PointerValue: // if reference object from external realm
			// XXX: consider pkgId here, A -> B - > A?...
			reo := tv.GetFirstObject(store)
			debug2.Println2("is reference to object, reo: ", reo)
			debug2.Println2("is reference to object, tv2.V, type of : ", tv2.V, reflect.TypeOf(tv2.V))

			// if a pointer has a base in
			// current realm, implies the
			// current realm is finalizing,
			// just skip as a recursive guard.
			if pv, ok := tv2.V.(PointerValue); ok {
				debug2.Println2("pv: ", pv)
				debug2.Println2("pv.TV: ", *pv.TV)
				// check recursive
				if b, ok := reo.(*Block); ok {
					if slices.Contains(b.Values, *pv.TV) { // this implies *pv.TV is real
						//if o2, ok := (*pv.TV).V.(Object); ok {
						//	if !o2.GetIsReal() {
						//  not return
						//	}
						//}
						debug2.Println2("return on block recursive")
						return
					}
				}
			}

			reo.SetOriginRealm(tv2.GetOriginPkg(store))
			reo.SetIsRef(true)
			checkCrossRealm(rlm, store, reo, true)
		}
	}
}

// checkCrossRealm performs a deep crawl to determine if cross-realm conditions exist.
// refValue is required to handle cases where the value is a slice.
// The `len` and `offset` are needed to validate proper elements of the underlying array.
// NOTE, oo can be real or unreal.
func checkCrossRealm(rlm *Realm, store Store, oo Object, isLastRef bool) {
	debug2.Println2("checkCrossRealm, oo: ", oo, reflect.TypeOf(oo))
	debug2.Printf2("isLastRef: %t, is current ref: %t \n", isLastRef, oo.GetIsRef())
	// is last not ref, current
	// object can be reference
	if !isLastRef {
		isLastRef = oo.GetIsRef()
	}

	if !oo.GetOriginRealm().IsZero() { // e.g. unreal array, struct...
		debug2.Println2("Origin Realm NOT zero...")
		if rlm.ID != oo.GetOriginRealm() { // crossing realm
			debug2.Println2("crossing realm, check oo, then elems")
			// reference value
			if isLastRef {
				debug2.Println2("Reference to object: ")
				if !oo.GetIsReal() {
					panic(fmt.Sprintf("cannot attach a reference to an unreal object from an external realm: %v", oo))
				} else {
					debug2.Println2("oo is real, just return")
					return
				}
			} else { // not reference to object
				debug2.Println2("Non reference object crossing realm, panic...")
				panic(fmt.Sprintf("cannot attach a value of a type defined by another realm: %v", oo))
			}
		} else {
			debug2.Println2("oo Not crossing realm, check elems...")
		}
	} else {
		debug2.Println2("Origin Realm is zero, unreal, check elems...")
	}

	switch v := oo.(type) {
	case *StructValue:
		// check fields
		for _, fv := range v.Fields {
			checkCrossRealm2(rlm, store, &fv, isLastRef) // ref to struct is heapItemValue or block
		}
	case *MapValue:
		debug2.Println2("MapValue, v: ", v)
		for cur := v.List.Head; cur != nil; cur = cur.Next {
			checkCrossRealm2(rlm, store, &cur.Key, isLastRef)
			checkCrossRealm2(rlm, store, &cur.Value, isLastRef)
		}
	case *BoundMethodValue:
	// TODO: complete this
	case *Block:
		debug2.Println2("BlockValue, v: ", v)
		debug2.Printf2("oo: %v, \n oo.GetRefCount: %v \n", oo, oo.GetRefCount())
		// NOTE, it's not escaped until now,
		// will set after check
		debug2.Println2("block is NOT real, v...PkgID: ", v.GetObjectID().PkgID)
		// TODO:, also check captures?
		for i, tv := range v.Values {
			debug2.Printf2("tv[%d] is tv: %v \n", i, tv)
			checkCrossRealm2(rlm, store, &tv, isLastRef)
		}
	case *HeapItemValue:
		if oo.GetRefCount() > 1 {
			debug2.Println2("hiv escaped, do nothing")
		} else {
			checkCrossRealm2(rlm, store, &v.Value, isLastRef)
		}
	case *ArrayValue:
		debug2.Println2("ArrayValue, v: ", v)
		for i, e := range v.List {
			debug2.Printf2("List, ArrayValue[%d] is %v: \n", i, e)
		}
		for i, e := range v.Data {
			debug2.Printf2("Data, ArrayValue[%d] is %v: \n", i, e)
		}
		//// TODO: return if it's real?
		//if oo.GetIsReal() {
		//	debug2.Println2("array IsReal, do nothing")
		//	return
		//} else {
		//	debug2.Println2("array is unreal")
		//}

		// if the array value is unreal,
		// it's going to be attached with
		// all the elems, so it's attaching
		// the array by value.
		if !oo.GetIsReal() {
			isLastRef = false
		}
		debug2.Println2("2, check elems of ArrayValue, v: ", v)
		if v.Data == nil {
			for _, e := range v.List {
				checkCrossRealm2(rlm, store, &e, isLastRef)
			}
		}
	default:
		panic("should not happen, oo is not object")
	}
}

//----------------------------------------
// mark*

// MarkNewEscapedCheckCrossRealm mark new escaped object
// and check cross realm
func (rlm *Realm) MarkNewEscapedCheckCrossRealm(store Store, oo Object) {
	debug2.Println2("MarkNewEscapedCheckCrossRealm, oo: ", oo)
	debug2.Println2("oo.GetOriginRealm(): ", oo.GetOriginRealm())
	debug2.Println2("isRef: ", oo.GetIsRef())
	debug2.Println2("rlm.ID: ", rlm.ID)

	if oo.GetOriginRealm() == rlm.ID {
		// do nothing
		return
	}

	if !oo.GetOriginRealm().IsZero() && oo.GetOriginRealm() != rlm.ID { // crossing realm
		if oo.GetIsRef() {
			checkCrossRealm(rlm, store, oo, oo.GetIsRef())
		} else {
			panic("cannot attach objects by value from external realm")
		}
	}
	// mark escaped
	if !oo.GetIsEscaped() {
		rlm.MarkNewEscaped(oo)
	}
}

func (rlm *Realm) MarkNewReal(oo Object) {
	debug2.Println2("MarkNewReal, oo:", oo)
	if debug {
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

// mark dirty == updated
func (rlm *Realm) MarkDirty(oo Object) {
	debug2.Println2("MarkDirty, oo: ", oo)
	if debug {
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
	if debug {
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
	if debug {
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

// TODO: check cross realm, that might be objects not attached
// to a realm gets attached here, which should panic.
// OpReturn calls this when exiting a realm transaction.
func (rlm *Realm) FinalizeRealmTransaction(readonly bool, store Store) {
	debug2.Println2("FinalizeRealmTransaction, rlm.ID: ", rlm.ID)
	defer func() {
		debug2.Println2("================done FinalizeRealmTransaction==================")
	}()
	if bm.OpsEnabled {
		bm.PauseOpCode()
		defer bm.ResumeOpCode()
	}
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
			panic("realm should not have created, deleted, or escaped marks before beginning finalization")
		}
	}
	// log realm boundaries in opslog.
	store.LogSwitchRealm(rlm.Path)
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
	debug2.Println2("processNewCreatedMarks")
	//fmt.Println("---len of newCreated objects:", len(rlm.newCreated))
	// Create new objects and their new descendants.
	//for _, oo := range rlm.newCreated {
	for i := 0; i < len(rlm.newCreated); i++ {
		oo := rlm.newCreated[i]
		debug2.Printf2("---oo[%d] is %v:\n", i, oo)
		//if _, ok := oo.(*BoundMethodValue); ok {
		//	panic("should not happen persist bound method")
		//}
		if debug {
			if oo.GetIsDirty() {
				panic("new created mark cannot be dirty")
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
			// check cross realm while attaching
			if pv, ok := oo.(*PackageValue); ok {
				checkCrossRealm(rlm, store, pv.Block.(Object), false)
			}
			rlm.incRefCreatedDescendants(store, oo)
		}
	}
	// Save new realm time.
	if len(rlm.newCreated) > 0 {
		store.SetPackageRealm(rlm)
	}
}

// XXX, unreal oo check happens in here
// oo must be marked new-real, and ref-count already incremented.
func (rlm *Realm) incRefCreatedDescendants(store Store, oo Object) {
	//debug2.Println2("---incRefCreatedDescendants from oo: ", oo, reflect.TypeOf(oo))
	//debug2.Println2("---oo.GetOriginRealm: ", oo.GetOriginRealm())
	//debug2.Println2("---oo.GetRefCount: ", oo.GetRefCount())
	//debug2.Println2("---rlm.ID: ", rlm.ID)
	//debug2.Println2("oo.GetOriginValue: ", oo.GetOriginValue())

	if debug {
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
		debug2.Println2("not zero, do nothing, return")
		return
	}
	rlm.assignNewObjectID(oo)
	rlm.created = append(rlm.created, oo) // XXX, here it becomes real.
	// RECURSE GUARD END

	// recurse for children.
	more := getChildObjects2(store, oo)
	for i, child := range more {
		debug2.Printf2("---[%d]child: %v, type of child: %v \n", i, child, reflect.TypeOf(child))
		if _, ok := child.(*PackageValue); ok {
			if debug {
				if child.GetRefCount() < 1 {
					panic("cannot increase reference count of package descendant that is unreferenced")
				}
			}
			// extern package values are skipped.
			continue
		}
		child.IncRefCount()
		rc := child.GetRefCount()
		debug2.Println2("rc after inc: ", rc)
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
		if debug {
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
	if debug {
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
			// do nothing
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
					panic("new escaped mark has no object ID")
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
	debug2.Println2("markDirtyAncestors")
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
					panic("ancestor should have a non-zero reference count to be marked as dirty")
				}
			}
			if rc > 1 {
				if debug {
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
	debug2.Println2("saveUnsavedObjects")
	for _, co := range rlm.created {
		debug2.Println2("co: ", co)
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
		debug2.Println2("uo: ", uo)
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
	debug2.Println2("saveUnsavedObjectRecursively", oo)
	if debug {
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
	// first, save unsaved children.
	unsaved := getUnsavedChildObjects(oo)
	debug2.Println2("unsaved: ", unsaved)
	for _, uch := range unsaved {
		debug2.Println2("uch: ", uch)
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
				panic("cannot save dirty new real object")
			}
		}
		rlm.saveObject(store, oo)
		oo.SetIsNewReal(false)
	} else {
		// update existing object.
		if debug {
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
	debug2.Println2("saveObject: ", oo)
	oid := oo.GetObjectID()
	//debug2.Println2("---oid: ", oid)
	//debug2.Println2("---oo.GetRefCount: ", oo.GetRefCount())
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
	store.SetObject(oo)
	// set index.
	if oo.GetIsEscaped() {
		// XXX save oid->hash to iavl.
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
				panic("cannot clear marks if new deleted exist")
			}
		}
		for _, oo := range rlm.newCreated {
			if oo.GetIsNewReal() {
				panic("cannot clear marks if new created exist")
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
		if bv, ok := cv.Closure.(*Block); ok {
			more = getSelfOrChildObjects(bv, more)
		}
		for _, c := range cv.Captures {
			more = getSelfOrChildObjects(c.V, more)
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
		//fmt.Println("block, cv: ", cv)
		//if _, ok := cv.Parent.(*Block); ok {
		//	fmt.Println("block, cv.parent: ", cv.Parent)
		//	fmt.Println("parent.Source: ", cv.Parent.(*Block).Source)
		//}
		for _, ctv := range cv.Values {
			more = getSelfOrChildObjects(ctv.V, more)
		}
		// Generally the parent block must also be persisted.
		// Otherwise NamePath may not resolve when referencing
		// a parent block.
		debug2.Println2("block value, get parent recursively")
		more = getSelfOrChildObjects(cv.Parent, more)
		return more
	case *HeapItemValue:
		more = getSelfOrChildObjects(cv.Value.V, more)
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
	debug2.Println2("getUnsavedChildObjects, val: ", val)
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
			PkgPath: ct.PkgPath,
			Methods: copyFieldsWithRefs(ct.Methods),
			Generic: ct.Generic,
		}
	case *TypeType:
		return &TypeType{}
	case *DeclaredType:
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
	case *NativeType:
		panic("cannot copy native types")
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
		if strings.HasSuffix(source.Location.File, "_test.gno") {
			// Ignore _test files
			return nil
		}
		var closure Value
		if cv.Closure != nil {
			closure = toRefValue(cv.Closure)
		}
		// nativeBody funcs which don't come from NativeResolver (and thus don't
		// have NativePkg/Name) can't be persisted, and should not be able
		// to get here anyway.
		if cv.nativeBody != nil && cv.NativePkg == "" {
			panic("cannot copy function value with native body when there is no native package")
		}
		ft := copyTypeWithRefs(cv.Type)
		return &FuncValue{
			Type:       ft,
			IsMethod:   cv.IsMethod,
			Source:     source,
			Name:       cv.Name,
			Closure:    closure,
			Captures:   cv.Captures,
			FileName:   cv.FileName,
			PkgPath:    cv.PkgPath,
			NativePkg:  cv.NativePkg,
			NativeName: cv.NativeName,
		}
	case *BoundMethodValue:
		fnc := copyValueWithRefs(cv.Func).(*FuncValue)
		rtv := refOrCopyValue(cv.Receiver)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(), // XXX ???
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
		// NOTE: While this could be eliminated sometimes with some
		// intelligence prior to persistence, to unwrap the
		// HeapItemValue in case where the HeapItemValue only has
		// refcount of 1,
		//
		//  1.  The HeapItemValue is necessary when the .Value is a
		//    primitive non-object anyways, and
		//  2. This would mean PointerValue.Base is nil, and we'd need
		//    additional logic to re-wrap when necessary, and
		//  3. And with the above point, it's not clear the result
		//    would be any faster.  But this is something we could
		//    explore after launch.
		hiv := &HeapItemValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Value:      refOrCopyValue(cv.Value),
		}
		return hiv
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

// (fully) fills the type.
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
		for i, mthd := range ct.Methods {
			ct.Methods[i].Type = fillType(store, mthd.Type)
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
	case *NativeType:
		panic("cannot fill native types")
	case blockType:
		return ct // nothing to do
	case *tupleType:
		for i, elt := range ct.Elts {
			ct.Elts[i] = fillType(store, elt)
		}
		return ct
	case RefType:
		return store.GetType(ct.TypeID())
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", reflect.TypeOf(typ)))
	}
}

func fillTypesTV(store Store, tv *TypedValue) {
	debug2.Println2("fillTypesTV, tv: ", tv)
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
		for i := 0; i < len(cv.List); i++ {
			ctv := &cv.List[i]
			fillTypesTV(store, ctv)
		}
		return cv
	case *SliceValue:
		fillTypesOfValue(store, cv.Base)
		return cv
	case *StructValue:
		for i := 0; i < len(cv.Fields); i++ {
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

			cv.vmap[cur.Key.ComputeMapKey(store, false)] = cur
		}
		return cv
	case TypeValue:
		cv.Type = fillType(store, cv.Type)
		return cv
	case *PackageValue:
		fillTypesOfValue(store, cv.Block)
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
	debug2.Printf2("assignNewObjectID, rlm: %v, oo: %v, oo: %p\n", rlm, oo, oo)
	oid := oo.GetObjectID()
	debug2.Println2("oid: ", oid)
	if !oid.IsZero() {
		panic("unexpected non-zero object id")
	}
	noid := rlm.nextObjectID()
	debug2.Println2("noid: ", noid)
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
					panic("cannot convert escaped object to ref value without an owner ID")
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
				panic("duplicate object")
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
