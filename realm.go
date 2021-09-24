package gno

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
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

// A nil realm is special and has limited functionality; enough
// to support methods that don't require persistence. This is
// the default realm when a machine starts with a non-realm
// package.  It could be said that pre-existing Go code runs in
// the nil realm and that no packages are realm packages.
type Realm struct {
	ID   RealmID
	Path string
	Time uint64

	created []Object  // new objects attached to real.
	updated []Object  // real objects that were modified.
	deleted []Object  // real objects that became deleted.
	ropslog []RealmOp // for debugging.
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

func (rlm *Realm) SetLogRealmOps(enabled bool) {
	if enabled {
		rlm.ResetRealmOps()
	} else {
		rlm.ropslog = nil
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
	if rlm == nil {
		return
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
		rlm.Time++
		oo.SetIsDirty(true, rlm.Time)
	}
	if rlm == nil {
		return
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
	if rlm == nil {
		return
	}
	//----------------------------------------
	// rlm != nil
	// append to .deleted
	if rlm.deleted == nil {
		rlm.deleted = make([]Object, 0, 256)
	}
	rlm.deleted = append(rlm.deleted, oo)
}

// removes deleted objects from created & updated.
func (rlm *Realm) CompressMarks() {

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

//----------------------------------------
// transactions

// OpReturn calls this when exiting a realm transaction.
// XXX don't need store because we don't need it when we save?
// XXX we need it when we load.
// XXX but this assumes no multiple references...
// XXX that is X owned by Y isn't also owned by Z.
// XXX unless we saved X but don't have reference to Y.
// XXX
func (rlm *Realm) FinalizeRealmTransaction(store Store) {
	// Process changes in created/updated/deleted.
	rlm.CompressMarks()
	rlm.ProcessCreatedObjects(store)
	rlm.ProcessUpdatedObjects(store)
	rlm.ProcessDeletedObjects(store)
	rlm.ClearMarks()
}

// crawls marked created objects and finalizes ownership
// by assigning it an ObjectID, recursively.
func (rlm *Realm) ProcessCreatedObjects(store Store) {
	for _, oo := range rlm.created {
		rlm.saveUnsavedObject(store, oo)
	}
}

func (rlm *Realm) saveUnsavedObject(store Store, oo Object) {
	if debug {
		if oo.GetIsProcessing() {
			panic("should not happen")
		}
		if oo.GetIsReal() && !oo.GetIsDirty() {
			panic("should not happen")
		}
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
		rlm.AssignNewObjectID(oo)
		// In case something loaded oo from disk (within the same
		// tx context), in-memory object oo must be linked for
		// identity mapping (w/ referential loops)..
		store.SetObject(oo)
	}
	// then process children
	more := getUnsavedChildren(oo, nil)
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
		rlm.saveUnsavedObject(store, child)
	}
	// save or update object
	if oo.GetIsNewReal() {
		rlm.SaveCreatedObject(oo)
	} else {
		rlm.SaveUpdatedObject(oo)
	}
}

// get unsaved self or unsaved children.
func getUnsaved(val Value, more []Object) []Object {
	if obj, ok := val.(Object); ok {
		if isUnsaved(obj) {
			return append(more, obj)
		} else {
			// nothing unsaved.
			return more
		}
	} else {
		return getUnsavedChildren(val, more)
	}
}

func getUnsavedChildren(val Value, more []Object) []Object {
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
			more = getUnsaved(cv.Base, more)
		} else {
			// If cv.Base is non-nil,
			// no need to append cv.Base's unsaved elements.
			more = getUnsaved(cv.TV.V, more)
		}
		return more
	case *ArrayValue:
		for _, ctv := range cv.List {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsaved(ctv.V, more)
		}
		return more
	case *SliceValue:
		more = getUnsaved(cv.Base, more)
		return more
	case *StructValue:
		for _, ctv := range cv.Fields {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsaved(ctv.V, more)
		}
		return more
	case *FuncValue:
		if bv, ok := cv.Closure.(*Block); ok {
			more = getUnsaved(bv, more)
		}
		return more
	case *BoundMethodValue:
		more = getUnsavedChildren(cv.Func, more)
		more = getUnsaved(cv.Receiver.V, more)
		return more
	case *MapValue:
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			// NOTE: same as isUnsaved(cur.Key.GetFirstObject()).
			more = getUnsaved(cur.Key.V, more)
			more = getUnsaved(cur.Value.V, more)
		}
		return more
	case TypeValue:
		return more
	case *PackageValue:
		for _, ctv := range cv.Values {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsaved(ctv.V, more)
		}
		for _, fb := range cv.FBlocks {
			more = getUnsaved(fb, more)
		}
		return more
	case *Block:
		for _, ctv := range cv.Values {
			// NOTE: same as isUnsaved(ctv.GetFirstObject()).
			more = getUnsaved(ctv.V, more)
		}
		more = getUnsaved(cv.Parent, more)
		return more
	case *nativeValue:
		return more // XXX ???
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
				Base:  ensureRefValue(parent, cv.Base),
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
			Base:   ensureRefValue(parent, cv.Base),
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
		var closure Value
		if cv.Closure != nil {
			closure = ensureRefValue(parent, cv.Closure)
		}
		if cv.nativeBody != nil {
			panic("should not happen")
		}
		ft := RefType{ID: cv.Type.TypeID()}
		return &FuncValue{
			Type:      ft,
			IsMethod:  cv.IsMethod,
			SourceLoc: cv.SourceLoc,
			Source:    cv.Source,
			Name:      cv.Name,
			Body:      cv.Body,
			Closure:   closure,
			FileName:  cv.FileName,
			PkgPath:   cv.PkgPath,
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
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = refOrCopy(cv, tv)
		}
		fblocks := make([]Value, len(cv.FBlocks))
		for i, fb := range cv.FBlocks {
			fblocks[i] = ensureRefValue(cv, fb)
		}
		return &PackageValue{
			Block: Block{
				ObjectInfo: cv.Block.ObjectInfo.Copy(),
				Source:     cv.Block.Source,
				Values:     vals,
				Parent:     nil,          // packages have no parent.
				Blank:      TypedValue{}, // empty
			},
			PkgName: cv.PkgName,
			PkgPath: cv.PkgPath,
			FNames:  cv.FNames, // no copy
			FBlocks: fblocks,
		}
	case *Block:
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = refOrCopy(cv, tv)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = ensureRefValue(parent, cv.Parent)
		}
		return &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     cv.Source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{}, // empty
		}
	case *nativeValue:
		return RefValue{}
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(val)))
	}
}

// crawls marked updated objects up the ownership chain
// to update the merkle hash.
func (rlm *Realm) ProcessUpdatedObjects(store Store) {
	for _, oo := range rlm.updated {
		rlm.saveUnsavedObject(store, oo)
	}
}

// crawls marked deleted objects, recursively.
func (rlm *Realm) ProcessDeletedObjects(store Store) {
	for _, do := range rlm.deleted {
		// Remove deleted object, and recursively
		// delete objects no longer referenced.
		rlm.RemoveDeletedObject(do)
	}
}

func (rlm *Realm) ClearMarks() {
	rlm.created = nil
	rlm.updated = nil
	rlm.deleted = nil
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
func (rlm *Realm) AssignNewObjectID(oo Object) ObjectID {
	oid := oo.GetObjectID()
	if !oid.IsZero() {
		panic("unexpected non-zero object id")
	}
	noid := rlm.nextObjectID()
	oo.SetObjectID(noid)
	oo.SetIsNewReal(true)
	return noid
}

func (rlm *Realm) SaveCreatedObject(oo Object) {
	rlm.saveObject(oo, RealmOpNew)
	oo.SetIsNewReal(false)
	oo.SetIsDirty(false, 0)
}

func (rlm *Realm) SaveUpdatedObject(oo Object) {
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
	rlm.saveObject(oo, RealmOpMod)
	oo.SetIsDirty(false, 0)
}

func (rlm *Realm) saveObject(oo Object, op RealmOpType) {
	oid := oo.GetObjectID()
	if oid.IsZero() {
		panic("unexpected zero object id")
	}
	// replace children/fields with Ref.
	o2 := copyWithRefs(nil, oo)
	// marshal to binary
	bz := amino.MustMarshal(o2)
	// set hash.
	hash := HashBytes(bz) // XXX objectHash(bz)???
	oo.SetHash(ValueHash{hash})
	// persist oid -> oo, bz(, hash???)
	rlm.saveObjectBytes(oid, bz, hash)
	// make realm op log entry
	if rlm.ropslog != nil {
		rlm.ropslog = append(rlm.ropslog,
			RealmOp{op, o2.(Object)})
	}
}

func (rlm *Realm) saveObjectBytes(oid ObjectID, bz []byte, hash Hashlet) {
	fmt.Println("XXX would save object bytes", oid) // , bz, hash)
}

func (rlm *Realm) RemoveDeletedObject(oo Object) {
	if rlm.ropslog != nil {
		rlm.ropslog = append(rlm.ropslog,
			RealmOp{RealmOpDel, oo})
	}
}

//----------------------------------------
// RealmOp
//
// At the end of a realm transaction, the operations
// are gathered into a buffer of RealmOps.

type RealmOpType uint8

const (
	RealmOpNew RealmOpType = iota
	RealmOpMod
	RealmOpDel
)

type RealmOp struct {
	Type RealmOpType
	Object
}

// used by the tests/file_test system to check
// veracity of realm operations.
func (rop RealmOp) String() string {
	switch rop.Type {
	case RealmOpNew:
		return fmt.Sprintf("c[%v]=%s",
			rop.Object.GetObjectID(),
			prettyJSON(amino.MustMarshalJSON(rop.Object)))
	case RealmOpMod:
		return fmt.Sprintf("u[%v]=%s",
			rop.Object.GetObjectID(),
			prettyJSON(amino.MustMarshalJSON(rop.Object)))
	case RealmOpDel:
		return fmt.Sprintf("d[%v]",
			rop.Object.GetObjectID())
	default:
		panic("should not happen")
	}
}

// resets .realmops.
func (rlm *Realm) ResetRealmOps() {
	rlm.ropslog = make([]RealmOp, 0, 1024)
}

// for test/file_test.go, to test realm changes.
func (rlm *Realm) SprintRealmOps() string {
	ss := make([]string, 0, len(rlm.ropslog))
	for _, rop := range rlm.ropslog {
		ss = append(ss, rop.String())
	}
	return strings.Join(ss, "\n")
}

//----------------------------------------
// Misc.

func ensureRefValue(parent Object, val Value) RefValue {
	if ref, ok := val.(RefValue); ok {
		return ref
	} else if oo, ok := val.(Object); ok {
		if !oo.GetIsReal() {
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
		tv.V = ensureRefValue(parent, obj)
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
