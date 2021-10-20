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
		NewTime: 0, // 0 reserved for package block.
	}
}

// A nil realm is special and has limited functionality; enough
// to support methods that don't require persistence. This is
// the default realm when a machine starts with a non-realm
// package.  It could be said that pre-existing Go code runs in
// the nil realm and that no packages are realm packages.
type Realm struct {
	ID   RealmID
	Path string
	Time uint64

	created []Object // new objects attached to real.
	updated []Object // real objects that were modified.
	deleted []Object // real objects that became deleted.
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
	if po == nil {
		return
	}
	if !po.GetIsReal() && !po.GetIsNewReal() {
		return // do nothing.
	}
	// NOTE: if po is a non-realm *PackageValue (which shouldn't happen
	// because they shouldn't have mutable state), the realm panics upon
	// getUnsavedObjects() during finalization.
	if co != nil {
		co.IncRefCount()
	}
	if xo != nil {
		xo.DecRefCount()
	}
	rlm.MarkDirty(po)
	if co != nil {
		if co.GetIsOwned() {
			if co.GetOwner() == po {
				// already owned by po but mark co as dirty
				// (refcount).  e.g. `a.bar = a.foo`
				if co.GetIsReal() {
					rlm.MarkDirty(co) // since refcount incremented
				}
			} else {
				// Owner conflict allowed within a transaction.  e.g. `b.foo
				// = a.foo; a.foo = nil` Conflicts will cause a panic upon
				// transaction finalization, when the owner's owned value's
				// OwnerID doesn't match the co's Owner's ID, or when
				// refcount isn't 1.  Corrolarily, there is no need to mark
				// the previous owner as dirty here.
				co.SetOwner(po)
				if co.GetIsReal() {
					rlm.MarkDirty(co) // since refcount incremented
				}
			}
		} else {
			co.SetOwner(po)
			rlm.MarkNewReal(co)
		}
	}
	if xo != nil {
		if xo.GetRefCount() == 0 {
			if debug {
				if xo.GetOwner() != po {
					panic("unexpected owner for deleted object")
				}
			}
			if xo.GetIsNewReal() || xo.GetIsReal() {
				rlm.MarkDeleted(xo)
			}
		}
	}
}

//----------------------------------------
// mark*

func (rlm *Realm) MarkNewReal(oo Object) {
	if debug {
		if oo.GetOwner() == nil {
			panic("should not happen")
		}
		if !oo.GetOwner().GetIsReal() {
			panic("should not happen")
		}
	}
	if oo.GetIsNewReal() {
		return // already marked.
	} else {
		oo.SetIsNewReal(true)
	}
	//----------------------------------------
	// rlm != nil
	// append to .created
	if rlm.created == nil {
		rlm.created = make([]Object, 0, 256)
	}
	rlm.created = append(rlm.created, oo)

}

func (rlm *Realm) MarkDirty(oo Object) {
	if debug {
		if !oo.GetIsReal() && !oo.GetIsNewReal() {
			panic("should not happen")
		}
	}
	if oo.GetIsDirty() {
		return // already marked.
	} else {
		oo.SetIsDirty(true, rlm.Time)
	}
	//----------------------------------------
	// rlm != nil
	// append to .updated
	if rlm.updated == nil {
		rlm.updated = make([]Object, 0, 256)
	}
	rlm.updated = append(rlm.updated, oo)
}

func (rlm *Realm) MarkDeleted(oo Object) {
	if debug {
		if !oo.GetIsNewReal() && !oo.GetIsReal() {
			panic("should not happen")
		}
		if oo.GetIsDeleted() {
			panic("should not happen")
		}
	}
	// NOTE: do not increment rlm.Time.
	// rlm.Time is passed in for debugging purposes.
	oo.SetIsDeleted(true, rlm.Time)
	//----------------------------------------
	// rlm != nil
	// append to .deleted
	if rlm.deleted == nil {
		rlm.deleted = make([]Object, 0, 256)
	}
	rlm.deleted = append(rlm.deleted, oo)
}

//----------------------------------------
// transactions

// OpReturn calls this when exiting a realm transaction.
// XXX don't need store because we don't need it when we save?
// XXX we need it when we load.
// XXX but this assumes no multiple references...
// XXX that is X owned by Y isn't also owned by Z.
// XXX unless we saved X but don't have reference to Y.
// XXX
func (rlm *Realm) FinalizeRealmTransaction(readonly bool, store Store) {
	// Process changes in created/updated/deleted.
	rlm.compressMarks()
	if readonly {
		if len(rlm.created) > 0 ||
			len(rlm.updated) > 0 ||
			len(rlm.deleted) > 0 {
			panic("realm updates in readonly transaction")
		}
	}
	rlm.processCreatedObjects(store)
	rlm.processUpdatedObjects(store)
	rlm.processDeletedObjects(store)
	rlm.clearMarks()
}

// removes deleted objects from created & updated.
func (rlm *Realm) compressMarks() {

	if debug {
		ensureUniq(rlm.created)
		ensureUniq(rlm.updated)
		ensureUniq(rlm.deleted)
	}

	c2 := make([]Object, 0, len(rlm.created))
	u2 := make([]Object, 0, len(rlm.updated))
	for _, co := range rlm.created {
		if co.GetIsDeleted() {
			// ignore deleted.
		} else {
			c2 = append(c2, co)
		}
	}
	for _, uo := range rlm.updated {
		if uo.GetIsDeleted() {
			// ignore deleted.
		} else {
			u2 = append(u2, uo)
		}
	}

	rlm.created = c2
	rlm.updated = u2
}

// crawls marked created objects and finalizes ownership
// by assigning it an ObjectID, recursively.
func (rlm *Realm) processCreatedObjects(store Store) {
	for _, oo := range rlm.created {
		rlm.saveUnsavedObjectRecursively(store, oo)
	}
}

// crawls marked updated objects up the ownership chain
// to update the merkle hash.
func (rlm *Realm) processUpdatedObjects(store Store) {
	for _, oo := range rlm.updated {
		rlm.saveUnsavedObjectRecursively(store, oo)
	}
}

// crawls marked deleted objects, recursively.
func (rlm *Realm) processDeletedObjects(store Store) {
	for _, do := range rlm.deleted {
		// Remove deleted object, and recursively
		// delete objects no longer referenced.
		rlm.removeDeletedObjectRecursively(store, do)
	}
}

func (rlm *Realm) clearMarks() {
	rlm.created = nil
	rlm.updated = nil
	rlm.deleted = nil
}

//----------------------------------------
// saveUnsavedObjectRecursively

func (rlm *Realm) saveUnsavedObjectRecursively(store Store, oo Object) {
	if debug {
		if oo.GetIsProcessing() {
			panic("should not happen")
		}
	}
	if oo.GetIsReal() && !oo.GetIsDirty() {
		return // already saved.
	}
	oo.SetIsProcessing(true)
	defer oo.SetIsProcessing(false)
	// first assign objectid if new
	if oo.GetObjectID().IsZero() {
		if debug {
			if oo.GetIsDirty() {
				panic("should not happen")
			}
		}
		rlm.assignNewObjectID(oo)
	}
	// then process children
	more := getUnsavedObjectsOfDescendants(oo, nil)
	for _, child := range more {
		if child.GetIsProcessing() {
			// Circular reference examples:
			// block -> fileblock -> fileblock.parent
			// They are OK for certain objects,
			// and are accounted for in RefCount.
			if child.GetObjectID().IsZero() {
				panic("should not happen")
			} else {
				break
			}
		}
		// XXX check for conflict? or before?
		child.SetOwner(oo)
		rlm.saveUnsavedObjectRecursively(store, child)
	}
	// save or update object
	if oo.GetIsNewReal() {
		rlm.saveCreatedObjectAndReset(store, oo)
	} else {
		rlm.saveUpdatedObjectAndReset(store, oo)
	}
}

// Get unsaved self or all unsaved descendants (deep).
func getUnsavedObjects(val Value, more []Object) []Object {
	// sanity check:
	if pv, ok := val.(*PackageValue); ok {
		if !pv.IsRealm() && pv.GetIsDirty() {
			panic("unexpected dirty non-realm package " + pv.PkgPath)
		}
	}
	if _, ok := val.(RefValue); ok {
		// ref means unchanged from disk.
		return more
	} else if obj, ok := val.(Object); ok {
		if isUnsaved(obj) {
			return append(more, obj)
		} else {
			// nothing unsaved.
			return more
		}
	} else {
		return getUnsavedObjectsOfDescendants(val, more)
	}
}

func getUnsavedObjectsOfDescendants(val Value, more []Object) []Object {
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
			more = getUnsavedObjects(cv.Base, more)
		} else {
			more = getUnsavedObjects(cv.TV.V, more)
		}
		return more
	case *ArrayValue:
		for _, ctv := range cv.List {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsavedObjects(ctv.V, more)
		}
		return more
	case *SliceValue:
		more = getUnsavedObjects(cv.Base, more)
		return more
	case *StructValue:
		for _, ctv := range cv.Fields {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsavedObjects(ctv.V, more)
		}
		return more
	case *FuncValue:
		if bv, ok := cv.Closure.(*Block); ok {
			more = getUnsavedObjects(bv, more)
		}
		return more
	case *BoundMethodValue:
		more = getUnsavedObjectsOfDescendants(cv.Func, more)
		more = getUnsavedObjects(cv.Receiver.V, more)
		return more
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			// NOTE: same as isUnsaved(cur.Key.GetFirstObject()).
			more = getUnsavedObjects(cur.Key.V, more)
			more = getUnsavedObjects(cur.Value.V, more)
		}
		return more
	case TypeValue:
		return more
	case *PackageValue:
		more = getUnsavedObjects(cv.Block, more)
		for _, fb := range cv.FBlocks {
			more = getUnsavedObjects(fb, more)
		}
		return more
	case *Block:
		for _, ctv := range cv.Values {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsavedObjects(ctv.V, more)
		}
		more = getUnsavedObjects(cv.Parent, more)
		return more
	case *NativeValue:
		panic("native values not supported")
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

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
	if tvv, ok := tv.V.(TypeValue); ok {
		more = append(more, tvv.Type)
	}
	return more
}

// Get unsaved types from a value.
// Unlike getUnsavedObjects(), only scans shallowly for types.
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

func fillType(store Store, ptr *Type) {
	if *ptr != nil {
		*ptr = store.GetType((*ptr).(RefType).TypeID())
	}
}

func fillTypesTV(store Store, tv *TypedValue) {
	if tv.T != nil {
		rt := tv.T.(RefType)
		tv.T = store.GetType(rt.TypeID())
	}
	if tvv, ok := tv.V.(TypeValue); ok {
		fillType(store, &tvv.Type)
		tv.V = tvv // since tvv is not addressable.
	}
}

// Partially fills loaded objects shallowly, similarly to getUnsavedTypes.
// Replaces all RefTypes with corresponding types.
func fillTypes(store Store, val Value) {
	switch cv := val.(type) {
	case nil: // do nothing
	case StringValue: // do nothing
	case BigintValue: // do nothing
	case DataByteValue: // do nothing
	case PointerValue:
		if cv.Base != nil {
			// cv.Base is object.
			// fillTypes(store, cv.Base) (wrong)
		} else {
			fillTypesTV(store, cv.TV)
		}
	case *ArrayValue:
		for i := 0; i < len(cv.List); i++ {
			ctv := &cv.List[i]
			fillTypesTV(store, ctv)
		}
	case *SliceValue:
		fillTypes(store, cv.Base)
	case *StructValue:
		for i := 0; i < len(cv.Fields); i++ {
			ctv := &cv.Fields[i]
			fillTypesTV(store, ctv)
		}
	case *FuncValue:
		fillType(store, &cv.Type)
		// XXX delete? (see GetUnsavedTypes()).
		//fillTypes(store, cv.Closure)
	case *BoundMethodValue:
		fillTypes(store, cv.Func)
		fillTypesTV(store, &cv.Receiver)
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			fillTypesTV(store, &cur.Key)
			fillTypesTV(store, &cur.Value)
		}
	case TypeValue: // do nothing
	case *PackageValue: // do nothing
	case *Block:
		for i := 0; i < len(cv.Values); i++ {
			ctv := &cv.Values[i]
			fillTypesTV(store, ctv)
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
// removeDeletedObjectRecursively

func (rlm *Realm) removeDeletedObjectRecursively(store Store, do Object) {
	// XXX actually delete objects recursively.
	store.DelObject(do)
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
	oo.SetIsNewReal(true)
	return noid
}

func (rlm *Realm) saveCreatedObjectAndReset(store Store, oo Object) {
	rlm.saveObject(store, oo)
	oo.SetIsNewReal(false)
	oo.SetIsDirty(false, 0)
}

func (rlm *Realm) saveUpdatedObjectAndReset(store Store, oo Object) {
	if debug {
		if oo.GetIsNewReal() {
			panic("should not happen")
		}
		if !oo.GetIsDirty() {
			panic("should not happen")
		}
		if !oo.GetIsReal() {
			panic("should not happen")
		}
	}
	rlm.saveObject(store, oo)
	oo.SetIsDirty(false, 0)
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
	// set object to store.
	store.SetObject(oo)
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
		if oo.GetRefCount() > 1 {
			parentID := parent.GetObjectID()
			if parentID == oo.GetOwnerID() {
				return RefValue{
					ObjectID: oo.GetObjectID(),
					Hash:     oo.GetHash(),
				}
			} else {
				return RefValue{
					ObjectID: oo.GetObjectID(),
					// Hash: nil,
				}
			}
		} else {
			return RefValue{
				ObjectID: oo.GetObjectID(),
				Hash:     oo.GetHash(),
			}
		}
	} else {
		panic("should not happen")
	}
}

func ensureUniq(ooz []Object) {
	om := make(map[Object]struct{}, len(ooz))
	for _, uo := range ooz {
		if _, ok := om[uo]; ok {
			panic("duplicate object")
		} else {
			om[uo] = struct{}{}
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
	return !oo.GetIsReal() || oo.GetIsDirty()
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
