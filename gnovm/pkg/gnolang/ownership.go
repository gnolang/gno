package gnolang

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

/*
## Ownership

In Gno, all objects are automatically persisted to disk after
every atomic "transaction" (a function call that must return
immediately.) when new objects are associated with a
"ownership tree" which is maintained overlaying the possibly
cyclic object graph (NOTE: cyclic references for persistence
not supported at this stage).  The ownership tree is composed
of objects (arrays, structs, maps, and blocks) and
derivatives (pointers, slices, and so on) with optional
struct-tag annotations to define the ownership tree.

If an object hangs off of the ownership tree, it becomes
included in the Merkle root, and is said to be "real".  The
Merkle-ized state of reality gets updated with state
transition transactions; during such a transaction, some new
temporary objects may "become real" by becoming associated in
the ownership tree (say, assigned to a struct field or
appended to a slice that was part of the ownership tree prior
to the transaction), but those that don't get garbage
collected and forgotten.

In the first release of Gno, all fields are owned in the same
realm, and no cyclic dependencies are allowed outside the
bounds of a realm transaction (this will change in phase 2,
where ref-counted references and weak references will be
supported).
*/

type ObjectID struct {
	PkgID   PkgID  // base
	NewTime uint64 // time created
}

func (oid ObjectID) MarshalAmino() (string, error) {
	pid := hex.EncodeToString(oid.PkgID.Hashlet[:])
	if oid.PkgID.purePkg {
		return fmt.Sprintf("%s:%s:%d", "purePkg", pid, oid.NewTime), nil
	} else {
		return fmt.Sprintf("%s:%d", pid, oid.NewTime), nil
	}
}

func (oid *ObjectID) UnmarshalAmino(oids string) error {
	parts := strings.Split(oids, ":")

	if len(parts) < 2 || len(parts) > 3 {
		return fmt.Errorf("invalid ObjectID %s", oids)
	}

	index1, index2 := 0, 1
	if len(parts) == 3 {
		index1, index2 = 1, 2
	}

	if parts[0] == "purePkg" {
		oid.PkgID.purePkg = true
	}

	_, err := hex.Decode(oid.PkgID.Hashlet[:], []byte(parts[index1]))
	if err != nil {
		return err
	}
	newTime, err := strconv.Atoi(parts[index2])
	if err != nil {
		return err
	}
	oid.NewTime = uint64(newTime)
	return nil
}

func (oid ObjectID) String() string {
	oids, _ := oid.MarshalAmino()
	return oids
}

// TODO: make faster by making PkgID a pointer
// and enforcing that the value of PkgID is never zero.
func (oid ObjectID) IsZero() bool {
	if debug {
		if oid.PkgID.IsZero() {
			if oid.NewTime != 0 {
				panic("should not happen")
			}
		}
	}
	return oid.PkgID.IsZero()
}

type Object interface {
	Value
	GetObjectInfo() *ObjectInfo
	GetObjectID() ObjectID
	MustGetObjectID() ObjectID
	SetObjectID(oid ObjectID)
	GetHash() ValueHash
	SetHash(ValueHash)
	GetOwner() Object
	GetOwnerID() ObjectID
	SetOwner(Object)
	GetIsOwned() bool
	// GetIsReal determines the reality of an Object.
	// During a transaction, the object is fake, but becomes real upon successful completion, making it persisted and verifiable.
	// This concept reflects a metaphysical understanding, where proof and persistence define an object's reality.
	GetIsReal() bool
	GetModTime() uint64
	IncRefCount() int
	DecRefCount() int
	GetRefCount() int
	GetIsDirty() bool
	SetIsDirty(bool, uint64)
	GetIsEscaped() bool
	SetIsEscaped(bool)
	GetIsDeleted() bool
	SetIsDeleted(bool, uint64)
	GetIsNewReal() bool
	SetIsNewReal(bool)
	GetBoundRealm() PkgID
	SetBoundRealm(pkgID PkgID)
	GetIsAttachingRef() bool
	SetIsAttachingRef(bool)
	GetIsNewEscaped() bool
	SetIsNewEscaped(bool)
	GetIsNewDeleted() bool
	SetIsNewDeleted(bool)
	GetIsTransient() bool

	// Saves to realm along the way if owned, and also (dirty
	// or new).
	// ValueImage(rlm *Realm, owned bool) *ValueImage
}

var (
	_ Object = &ArrayValue{}
	_ Object = &StructValue{}
	_ Object = &BoundMethodValue{}
	_ Object = &MapValue{}
	_ Object = &Block{}
	_ Object = &HeapItemValue{}
)

type ObjectInfo struct {
	ID       ObjectID  // set if real.
	Hash     ValueHash `json:",omitempty"` // zero if dirty.
	OwnerID  ObjectID  `json:",omitempty"` // parent in the ownership tree.
	ModTime  uint64    // time last updated.
	RefCount int       // for persistence. deleted/gc'd if 0.

	// Object has multiple references (refcount > 1) and is persisted separately
	IsEscaped bool `json:",omitempty"` // hash in iavl.

	// MemRefCount int // consider for optimizations.

	// Object has been modified and needs to be saved
	isDirty bool

	// Object has been permanently deleted
	isDeleted bool

	// Object is newly created in current transaction and will be persisted
	isNewReal bool

	// Object newly created multiple references in current transaction
	isNewEscaped bool

	// Object is marked for deletion in current transaction
	isNewDeleted bool

	// realm where object is from
	boundRealm PkgID

	// This flag indicates whether the object is being
	// attached as a base of reference or as itself.
	// For example:
	// - If the object being attached is a struct value
	// whose type is declared in another realm, it should panic.
	// - If the object being attached is a pointer to such
	// a struct value, it is allowed.
	isAttachingRef bool

	// XXX huh?
	owner Object // mem reference to owner.
}

// Copy used for serialization of objects.
// Note that "owner" is nil.
func (oi *ObjectInfo) Copy() ObjectInfo {
	return ObjectInfo{
		ID:        oi.ID,
		Hash:      oi.Hash.Copy(),
		OwnerID:   oi.OwnerID,
		ModTime:   oi.ModTime,
		RefCount:  oi.RefCount,
		IsEscaped: oi.IsEscaped,
	}
}

func (oi *ObjectInfo) String() string {
	// XXX update with new flags.
	return fmt.Sprintf(
		"OI[%s#%X,owner=%s,refs=%d,new:%v,drt:%v,del:%v]",
		oi.ID.String(),
		oi.Hash.Bytes(),
		oi.OwnerID.String(),
		oi.RefCount,
		oi.GetIsNewReal(),
		oi.GetIsDirty(),
		oi.GetIsDeleted(),
	)
}

func (oi *ObjectInfo) GetObjectInfo() *ObjectInfo {
	return oi
}

func (oi *ObjectInfo) GetObjectID() ObjectID {
	return oi.ID
}

func (oi *ObjectInfo) MustGetObjectID() ObjectID {
	if oi.ID.IsZero() {
		panic("unexpected zero object id")
	}
	return oi.ID
}

func (oi *ObjectInfo) SetObjectID(oid ObjectID) {
	oi.ID = oid
}

func (oi *ObjectInfo) GetHash() ValueHash {
	return oi.Hash
}

func (oi *ObjectInfo) SetHash(vh ValueHash) {
	oi.Hash = vh
}

func (oi *ObjectInfo) GetOwner() Object {
	return oi.owner
}

func (oi *ObjectInfo) SetOwner(po Object) {
	if po == nil {
		oi.OwnerID = ObjectID{}
		oi.owner = nil
	} else {
		oi.OwnerID = po.GetObjectID()
		oi.owner = po
	}
}

func (oi *ObjectInfo) GetOwnerID() ObjectID {
	//if oi.owner == nil {
	//	return ObjectID{}
	//} else {
	//	return oi.owner.GetObjectID()
	//}
	return oi.OwnerID
}

func (oi *ObjectInfo) GetIsOwned() bool {
	return !oi.OwnerID.IsZero()
}

// NOTE: does not return true for new reals.
func (oi *ObjectInfo) GetIsReal() bool {
	return !oi.ID.IsZero()
}

func (oi *ObjectInfo) GetModTime() uint64 {
	return oi.ModTime
}

func (oi *ObjectInfo) IncRefCount() int {
	oi.RefCount++
	return oi.RefCount
}

func (oi *ObjectInfo) DecRefCount() int {
	oi.RefCount--
	if oi.RefCount < 0 {
		// This may happen for uninitialized values.
		if debug {
			if oi.GetIsReal() {
				panic("should not happen")
			}
		}
	}
	return oi.RefCount
}

func (oi *ObjectInfo) GetRefCount() int {
	return oi.RefCount
}

func (oi *ObjectInfo) GetIsDirty() bool {
	return oi.isDirty
}

func (oi *ObjectInfo) SetIsDirty(x bool, mt uint64) {
	if x {
		oi.Hash = ValueHash{}
		oi.ModTime = mt
	}
	oi.isDirty = x
}

func (oi *ObjectInfo) GetIsEscaped() bool {
	return oi.IsEscaped
}

func (oi *ObjectInfo) SetIsEscaped(x bool) {
	oi.IsEscaped = x
}

func (oi *ObjectInfo) GetIsDeleted() bool {
	return oi.isDeleted
}

func (oi *ObjectInfo) SetIsDeleted(x bool, mt uint64) {
	// NOTE: Don't over-write modtime.
	// Consider adding a DelTime, or just log it somewhere, or
	// continue to ignore it.

	// The above comment is likely made because it could introduce complexity
	// Objects can be "undeleted" if referenced during a transaction
	// If an object is deleted and then undeleted in the same transaction
	// If an object is deleted multiple times
	// ie...continue to ignore it
	oi.isDeleted = x
}

func (oi *ObjectInfo) GetIsNewReal() bool {
	return oi.isNewReal
}

func (oi *ObjectInfo) SetIsNewReal(x bool) {
	oi.isNewReal = x
}

func (oi *ObjectInfo) GetBoundRealm() PkgID {
	return oi.boundRealm
}

func (oi *ObjectInfo) SetBoundRealm(pkgId PkgID) {
	oi.boundRealm = pkgId
}

func (oi *ObjectInfo) GetIsAttachingRef() bool {
	return oi.isAttachingRef
}

func (oi *ObjectInfo) SetIsAttachingRef(ref bool) {
	oi.isAttachingRef = ref
}

func (oi *ObjectInfo) GetIsNewEscaped() bool {
	return oi.isNewEscaped
}

func (oi *ObjectInfo) SetIsNewEscaped(x bool) {
	oi.isNewEscaped = x
}

func (oi *ObjectInfo) GetIsNewDeleted() bool {
	return oi.isNewDeleted
}

func (oi *ObjectInfo) SetIsNewDeleted(x bool) {
	oi.isNewDeleted = x
}

func (oi *ObjectInfo) GetIsTransient() bool {
	return false
}

// get first accessible object, maybe containing(parent) object, maybe itself.
func (tv *TypedValue) GetFirstObject(store Store) Object {
	switch cv := tv.V.(type) {
	case PointerValue:
		return cv.GetBase(store)
	case *ArrayValue:
		return cv
	case *SliceValue:
		return cv.GetBase(store)
	case *StructValue:
		return cv
	case *FuncValue:
		return cv.GetClosure(store)
	case *MapValue:
		return cv
	case *BoundMethodValue:
		return cv
	case *NativeValue:
		// XXX allow PointerValue.Assign2 to pass nil for oo1/oo2.
		// panic("realm logic for native values not supported")
		return nil
	case *Block:
		return cv
	case RefValue:
		oo := store.GetObject(cv.ObjectID)
		tv.V = oo
		return oo
	case *HeapItemValue:
		// should only appear in PointerValue.Base
		panic("heap item value should only appear as a pointer's base")
	default:
		return nil
	}
}

func (tv *TypedValue) GetFirstObject2(store Store) Object {
	//fmt.Println("GetFirstObject2, tv: ", tv, reflect.TypeOf(tv.V))
	obj := tv.GetFirstObject(store)

	// infer original package using declared type
	getPkgId := func(t Type) (pkgId PkgID) {
		if dt, ok := t.(*DeclaredType); ok {
			pkgId = PkgIDFromPkgPath(dt.GetPkgPath())
		}
		return
	}

	var originPkg PkgID

	switch cv := obj.(type) {
	case *HeapItemValue:
		originPkg = getPkgId(cv.Value.T)
	case *Block:
		// assert to pointer value
		if pv, ok := tv.V.(PointerValue); ok {
			originPkg = getPkgId(pv.TV.T)
		} else {
			// XXX?
		}
	case *BoundMethodValue:
		// do nothing
	case *MapValue, *StructValue, *ArrayValue:
		// if it's a declared type, origin realm
		// is deduced from type, otherwise zero.
		originPkg = getPkgId(tv.T)
	default:
		// do nothing
	}

	// set origin realm to object
	if obj != nil && !originPkg.IsZero() {
		// attach bound package info
		// used for checking cross realm after
		obj.SetBoundRealm(originPkg)
		switch tv.V.(type) {
		case *SliceValue, PointerValue:
			//fmt.Println("match!!!")
			//fmt.Println("real? ", obj.GetIsReal())
			//obj.SetIsAttachingRef(true)
			if obj.GetIsReal() { // if not real, is attaching by value, e.g. heapItemValue
				obj.SetIsAttachingRef(true)
			}
		}
	}
	return obj
}

// GetBoundRealmByType retrieves the bound realm for the object
// by checking its type. If the type is a declared type, it is
// considered bound to a specific realm; otherwise, it returns zero.
func (tv *TypedValue) GetBoundRealmByType(obj Object) (originPkg PkgID) {
	// infer original package using declared type
	getPkgId := func(t Type) (pkgId PkgID) {
		if dt, ok := t.(*DeclaredType); ok {
			pkgId = PkgIDFromPkgPath(dt.GetPkgPath())
		}
		return
	}

	switch cv := obj.(type) {
	case *HeapItemValue:
		originPkg = getPkgId(cv.Value.T)
		return
	case *Block:
		// assert to pointer value
		if pv, ok := tv.V.(PointerValue); ok {
			originPkg = getPkgId(pv.TV.T)
			return
		} else {
			// XXX?
		}
	case *BoundMethodValue:
		// do nothing
		return
	case *MapValue, *StructValue, *ArrayValue:
		// if it's a declared type, origin realm
		// is deduced from type, otherwise zero.
		originPkg = getPkgId(tv.T)
		return
	default:
		// do nothing
	}
	return
}
