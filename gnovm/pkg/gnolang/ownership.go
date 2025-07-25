package gnolang

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
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
	return fmt.Sprintf("%s:%d", pid, oid.NewTime), nil
}

func (oid *ObjectID) UnmarshalAmino(oids string) error {
	parts := strings.Split(oids, ":")
	if len(parts) != 2 {
		return errors.New("invalid ObjectID %s", oids)
	}
	_, err := hex.Decode(oid.PkgID.Hashlet[:], []byte(parts[0]))
	if err != nil {
		return err
	}
	newTime, err := strconv.Atoi(parts[1])
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

type ObjectIDer interface {
	GetObjectID() ObjectID
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
	GetIsNewEscaped() bool
	SetIsNewEscaped(bool)
	GetIsNewDeleted() bool
	SetIsNewDeleted(bool)
	GetIsTransient() bool

	GetLastGCCycle() int64
	SetLastGCCycle(int64)

	// Saves to realm along the way if owned, and also (dirty
	// or new).
	// ValueImage(rlm *Realm, owned bool) *ValueImage
}

var (
	_ Object = &ArrayValue{}
	_ Object = &StructValue{}
	_ Object = &FuncValue{}
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

	LastObjectSize int64 //

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
	lastGCCycle  int64
	owner        Object // mem reference to owner.
}

// Copy used for serialization of objects.
// Note that "owner" is nil.
func (oi *ObjectInfo) Copy() ObjectInfo {
	return ObjectInfo{
		ID:             oi.ID,
		Hash:           oi.Hash.Copy(),
		OwnerID:        oi.OwnerID,
		ModTime:        oi.ModTime,
		RefCount:       oi.RefCount,
		IsEscaped:      oi.IsEscaped,
		LastObjectSize: oi.LastObjectSize,
		isDirty:        oi.isDirty,
		isDeleted:      oi.isDeleted,
		isNewReal:      oi.isNewReal,
		isNewEscaped:   oi.isNewEscaped,
		isNewDeleted:   oi.isNewDeleted,
		lastGCCycle:    oi.lastGCCycle,
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
	if oi.owner == nil {
		return ObjectID{}
	} else {
		return oi.owner.GetObjectID()
	}
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

func (oi *ObjectInfo) GetLastGCCycle() int64 {
	return oi.lastGCCycle
}

func (oi *ObjectInfo) SetLastGCCycle(c int64) {
	oi.lastGCCycle = c
}

func (oi *ObjectInfo) GetIsTransient() bool {
	return false
}

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
		return cv
	case *MapValue:
		return cv
	case *BoundMethodValue:
		return cv
	case *PackageValue:
		return cv.GetBlock(store)
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

// returns false if there is no object.
func (tv *TypedValue) GetFirstObjectID() (ObjectID, bool) {
	switch cv := tv.V.(type) {
	case PointerValue:
		if cv.Base == nil {
			return ObjectID{}, false
		}
		return cv.Base.(ObjectIDer).GetObjectID(), true
	case *ArrayValue:
		return cv.GetObjectID(), true
	case *SliceValue:
		return cv.Base.(ObjectIDer).GetObjectID(), true
	case *StructValue:
		return cv.GetObjectID(), true
	case *FuncValue:
		return cv.GetObjectID(), true
	case *MapValue:
		return cv.GetObjectID(), true
	case *BoundMethodValue:
		return cv.GetObjectID(), true
	case *PackageValue:
		return cv.Block.(ObjectIDer).GetObjectID(), true
	case *Block:
		return cv.GetObjectID(), true
	case RefValue:
		return cv.GetObjectID(), true
	case *HeapItemValue:
		// should only appear in PointerValue.Base
		panic("heap item value should only appear as a pointer's base")
	default:
		return ObjectID{}, false
	}
}
