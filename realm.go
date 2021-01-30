package gno

import (
	"fmt"
	"strings"
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

type Realm struct {
	ID      RealmID
	Path    string
	Counter uint64

	created []Object      // new objects attached to real.
	updated []Object      // real objects that were modified.
	deleted []Object      // real objects that became deleted.
	ropslog []RealmOp     // for debugging.
	pkg     *PackageValue // associated package if any.
}

// Creates a blank new realm with counter 0.
func NewRealm(path string) *Realm {
	return &Realm{
		ID:      RealmIDFromPath(path),
		Path:    path,
		Counter: 0,
	}
}

func (rlm *Realm) String() string {
	return fmt.Sprintf("Realm{Path:%q:Counter:%d}#%X",
		rlm.Path, rlm.Counter, rlm.ID.Bytes())
}

func (rlm *Realm) SetLogRealmOps(enabled bool) {
	if enabled {
		rlm.ropslog = make([]RealmOp, 0, 1024)
	} else {
		rlm.ropslog = nil
	}
}

// object co attached to po
func (rlm *Realm) DidAttachTo(co, po Object) {
	if debug {
		if po == nil {
			panic("should not happen")
		}
		if co.GetIsDeleted() {
			panic("cannot attach a deleted object")
		}
		if po.GetIsDeleted() {
			panic("cannot attach to a deleted object")
		}
	}
	co.IncRefCount()
	if !co.GetIsOwned() {
		co.SetOwner(po)
		if po.GetIsReal() {
			rlm.MarkNewReal(co)
			rlm.MarkDirty(po)
		}
	} else if co.GetOwner() == po {
		// already owned by po but inc owncount and mark dirty.
		// e.g. `a.bar = a.foo`
		if po.GetIsReal() {
			rlm.MarkDirty(po)
		}
	} else {
		// owner conflict allowed within a transaction.
		// e.g. `b.foo = a.foo; a.foo = nil`
		// conflicts will cause a panic upon transaction finalization,
		// when owner's owned value doesn't match co's Owner.
		ex := co.GetOwner()
		co.SetOwner(po)
		if ex.GetIsReal() {
			rlm.MarkDirty(ex)
		}
		if po.GetIsReal() {
			rlm.MarkDirty(po)
		}
	}
}

func (rlm *Realm) DidUpdate(oo Object) {
	if debug {
		if oo.GetIsDeleted() {
			panic("cannot update to a deleted object")
		}
	}
	if oo.GetIsReal() {
		fmt.Println("QQQ")
		rlm.MarkDirty(oo)
	}
}

func (rlm *Realm) DidDetach(oo Object) {
	if debug {
		if oo.GetOwner() == nil {
			panic("should not happen")
		}
		if oo.GetIsDeleted() {
			panic("cannot delete a deleted object")
		}
	}
	ex := oo.GetOwner()
	oo.SetOwner(nil)
	if ex.GetIsReal() {
		rlm.MarkDirty(ex)
	}
	if oo.DecRefCount() == 0 {
		if oo.GetIsNewReal() || oo.GetIsReal() {
			rlm.MarkDeleted(oo)
		}
	}
}

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
		oo.SetIsDirty(true)
	}
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
	oo.SetIsDeleted(true)
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
	// XXX actually create
}

// crawls marked updated objects up the ownership chain
// to update the merkle hash.
func (rlm *Realm) ProcessUpdatedObjects() {
	for _, uo := range rlm.updated {
		// XXX actually update.
		if rlm.ropslog != nil {
			rlm.ropslog = append(rlm.ropslog,
				RealmOp{
					Type:   RealmOpMod,
					Object: uo,
				})
		}
	}
}

// crawls marked deleted objects, recursively.
func (rlm *Realm) ProcessDeletedObjects() {
	// XXX actually delete.
}

func (rlm *Realm) ClearMarks() {
	rlm.created = nil
	rlm.updated = nil
	rlm.deleted = nil
}

//----------------------------------------

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
}

// used by the tests/file_test system to check
// veracity of realm operations.
func (rop RealmOp) String() string {
	switch rop.Type {
	case RealmOpNew:
		return "NOTYETIMPL"
	case RealmOpMod:
		// NOTE: assumes *Realm is no longer needed.
		var rlm *Realm = nil
		return fmt.Sprintf("u[%v]=%v",
			rop.Object.GetObjectID(),
			rop.Object.ValuePreimage(rlm, true).String())
	case RealmOpDel:
		return "NOTYETIMPL"
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
