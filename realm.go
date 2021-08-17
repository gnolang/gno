package gno

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// Realm

type RealmID struct {
	Hashlet
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

type Realmer func(pkgPath string) *Realm

// A nil realm is special and has limited functionality; enough
// to support methods that don't require persistence. This is
// the default realm when a machine starts with a non-realm
// package.  It could be said that pre-existing Go code runs in
// the nil realm and that no packages are realm packages.
type Realm struct {
	ID   RealmID
	Path string
	Time uint64
	ImageCodec

	created []Object      // new objects attached to real.
	updated []Object      // real objects that were modified.
	deleted []Object      // real objects that became deleted.
	ropslog []RealmOp     // for debugging.
	pkg     *PackageValue // associated package if any.
}

// Creates a blank new realm with counter 0.
func NewRealm(path string) *Realm {
	id := RealmIDFromPath(path)
	return &Realm{
		ID:   id,
		Path: path,
		Time: 0,
		ImageCodec: ImageCodec{
			RealmID:       id,
			TypeLookup:    nil,
			PackageLookup: nil,
		},
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
		rlm.ropslog = make([]RealmOp, 0, 1024)
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
	if co != nil {
		co.IncRefCount()
	}
	if xo != nil {
		xo.DecRefCount()
	}
	if po == nil {
		return
	}
	if !po.GetIsReal() {
		// Object may become new-real after tx if it is
		// indirectly owned by something real.  We don't
		// know yet, but we will mark it later when we do
		// after assigning it an ObjectID()..
		//
		// Also, if po isn't real, don't bother to mark it
		// dirty, since it will already become marked as
		// new-real and get saved anyways if it is  real
		// post tx.
		return // do nothing.
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
		if !oo.GetIsReal() {
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
func (rlm *Realm) FinalizeRealmTransaction() {
	// Process changes in created/updated/deleted.
	rlm.CompressMarks()
	rlm.ProcessCreatedObjects()
	rlm.ProcessUpdatedObjects()
	rlm.ProcessDeletedObjects()
	rlm.ClearMarks()
}

// crawls marked created objects and finalizes ownership
// by assigning it an ObjectID, recursively.
func (rlm *Realm) ProcessCreatedObjects() {
	for _, oo := range rlm.created {
		rlm.processCreatedOrUpdatedObject(oo)
	}
}

func (rlm *Realm) processCreatedOrUpdatedObject(oo Object) {
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
	// first process children
	more := getCreatedOrUpdatedChildren(oo)
	for _, child := range more {
		if child.GetIsProcessing() {
			// NOTE: circular references not yet supported.
			panic("should not happen")
		} else {
			rlm.processCreatedOrUpdatedObject(child)
		}
	}
	// save or update object
	if oo.GetIsDirty() {
		rlm.SaveUpdatedObject(oo)
	} else {
		rlm.SaveCreatedObject(oo)
	}
}

func getCreatedOrUpdatedChildren(obj Object) []Object {
	switch co := obj.(type) {
	case *ArrayValue:
		more := make([]Object, 0, len(co.List))
		for _, ctv := range co.List {
			if cobj, ok := ctv.V.(Object); ok {
				if !cobj.GetIsReal() {
					more = append(more, cobj)
				} else if cobj.GetIsDirty() {
					more = append(more, cobj)
				}
			}
		}
		return more
	case *StructValue:
		more := make([]Object, 0, len(co.Fields))
		for _, ctv := range co.Fields {
			if cobj, ok := ctv.V.(Object); ok {
				if !cobj.GetIsReal() {
					more = append(more, cobj)
				} else if cobj.GetIsDirty() {
					more = append(more, cobj)
				}
			}
		}
		return more
	case *MapValue:
		more := make([]Object, 0, 2*co.List.Size)
		for cur := co.List.Head; cur != nil; cur = cur.Next {
			if cobj, ok := cur.Key.V.(Object); ok {
				if !cobj.GetIsReal() {
					more = append(more, cobj)
				} else if cobj.GetIsDirty() {
					more = append(more, cobj)
				}
			}
			if cobj, ok := cur.Value.V.(Object); ok {
				if !cobj.GetIsReal() {
					more = append(more, cobj)
				} else if cobj.GetIsDirty() {
					more = append(more, cobj)
				}
			}
		}
		return more
	case *Block:
		more := make([]Object, 0, len(co.Values))
		for _, ctv := range co.Values {
			if cobj, ok := ctv.V.(Object); ok {
				if !cobj.GetIsReal() {
					more = append(more, cobj)
				} else if cobj.GetIsDirty() {
					more = append(more, cobj)
				}
			}
		}
		return more
	default:
		panic(fmt.Sprintf(
			"unexpected type %v",
			reflect.TypeOf(obj)))
	}
}

// crawls marked updated objects up the ownership chain
// to update the merkle hash.
func (rlm *Realm) ProcessUpdatedObjects() {
	for _, oo := range rlm.updated {
		rlm.processCreatedOrUpdatedObject(oo)
	}
}

// crawls marked deleted objects, recursively.
func (rlm *Realm) ProcessDeletedObjects() {
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
func (rlm *Realm) AssignObjectID(oo Object) ObjectID {
	oid := oo.GetObjectID()
	if !oid.IsZero() {
		panic("unexpected non-zero object id")
	}
	noid := rlm.nextObjectID()
	oo.SetObjectID(noid)
	oo.SetIsNewReal(true) // TODO remove?
	return noid
}

func (rlm *Realm) SaveCreatedObject(oo Object) {
	rlm.AssignObjectID(oo)
	oi := rlm.EncodeObjectImage(oo)
	oo.SetHash(hashValueImage(oi))
	rlm.saveObject(oo, oi)
	if rlm.ropslog != nil {
		rlm.ropslog = append(rlm.ropslog,
			RealmOp{RealmOpNew, oo, oi})
	}
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
	oi := rlm.EncodeObjectImage(oo)
	oo.SetHash(hashValueImage(oi))
	rlm.saveObject(oo, oi)
	if rlm.ropslog != nil {
		rlm.ropslog = append(rlm.ropslog,
			RealmOp{RealmOpMod, oo, oi})
	}
	oo.SetIsDirty(false, 0)
}

func (rlm *Realm) saveObject(oo Object, vi ValueImage) {
	oid := oo.GetObjectID()
	if oid.IsZero() {
		panic("unexpected zero object id")
	}
	fmt.Printf("XXX WOULD SAVE: %v=%v\n", oid, vi)
}

func (rlm *Realm) RemoveDeletedObject(oo Object) {
	if rlm.ropslog != nil {
		rlm.ropslog = append(rlm.ropslog,
			RealmOp{RealmOpDel, oo, nil})
	}
}

//----------------------------------------
// misc

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

func IsRealmPath(pkgPath string) bool {
	// TODO: make it more distinct to distinguish from normal paths.
	if strings.HasPrefix(pkgPath, "gno.land/r/") {
		return true
	} else {
		return false
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
	ValueImage
}

// used by the tests/file_test system to check
// veracity of realm operations.
func (rop RealmOp) String() string {
	switch rop.Type {
	case RealmOpNew:
		return fmt.Sprintf("c[%v]=%v",
			rop.Object.GetObjectID(),
			spew.Sdump(rop.ValueImage))
	case RealmOpMod:
		return fmt.Sprintf("u[%v]=%v",
			rop.Object.GetObjectID(),
			spew.Sdump(rop.ValueImage))
	case RealmOpDel:
		return fmt.Sprintf("d[%v]",
			rop.Object.GetObjectID())
	default:
		panic("should not happen")
	}
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
// MemRealmer

func NewMemRealmer() Realmer {
	rlms := make(map[string]*Realm)
	return Realmer(func(pkgPath string) *Realm {
		if !IsRealmPath(pkgPath) {
			panic("should not happen")
		}
		if rlm, ok := rlms[pkgPath]; ok {
			return rlm
		} else {
			rlm = NewRealm(pkgPath)
			rlms[pkgPath] = rlm
			return rlm
		}
	})
}
