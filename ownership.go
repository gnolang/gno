package gno

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/pkgs/errors"
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
	RealmID RealmID // base
	NewTime uint64  // time created
}

func (oid ObjectID) MarshalAmino() (string, error) {
	rid := hex.EncodeToString(oid.RealmID.Hashlet[:])
	return fmt.Sprintf("%s:%d", rid, oid.NewTime), nil
}

func (oid *ObjectID) UnmarshalAmino(oids string) error {
	parts := strings.Split(oids, ":")
	if len(parts) != 2 {
		return errors.New("invalid ObjectID %s", oids)
	}
	_, err := hex.Decode(oid.RealmID.Hashlet[:], []byte(parts[0]))
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

func (oid ObjectID) Bytes() []byte {
	bz := make([]byte, HashSize+8)
	copy(bz[:HashSize], oid.RealmID.Bytes())
	binary.BigEndian.PutUint64(
		bz[HashSize:], uint64(oid.NewTime))
	return bz
}

// TODO: make faster by making RealmID a pointer
// and enforcing that the value of RealmID is never zero.
func (oid ObjectID) IsZero() bool {
	if debug {
		if oid.RealmID.IsZero() && oid.NewTime != 0 {
			panic("should not happen")
		}
	}
	return oid.RealmID.IsZero()
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
	GetIsReal() bool
	GetModTime() uint64
	IncRefCount() int
	DecRefCount() int
	GetRefCount() int
	GetIsNewReal() bool
	SetIsNewReal(bool)
	GetIsDirty() bool
	SetIsDirty(bool, uint64)
	GetIsDeleted() bool
	SetIsDeleted(bool, uint64)
	GetIsProcessing() bool
	SetIsProcessing(bool)
	GetIsTransient() bool

	// Saves to realm along the way if owned, and also (dirty
	// or new).
	// ValueImage(rlm *Realm, owned bool) *ValueImage
}

var _ Object = &ArrayValue{}
var _ Object = &StructValue{}
var _ Object = &BoundMethodValue{}
var _ Object = &MapValue{}
var _ Object = &Block{}

type ObjectInfo struct {
	ID       ObjectID  // set if real.
	Hash     ValueHash `json:",omitempty"` // zero if dirty.
	OwnerID  ObjectID  `json:",omitempty"` // parent in the ownership tree.
	ModTime  uint64    // time last updated.
	RefCount int       // for persistence. deleted/gc'd if 0.
	// MemRefCount int // consider for optimizations.
	isNewReal    bool
	isDirty      bool
	isDeleted    bool
	isProcessing bool

	// XXX huh?
	owner Object // mem reference to owner.
}

// Copy used for serialization of objects.
// Note that "owner" is nil.
func (oi *ObjectInfo) Copy() ObjectInfo {
	return ObjectInfo{
		ID:           oi.ID,
		Hash:         oi.Hash.Copy(),
		OwnerID:      oi.OwnerID,
		ModTime:      oi.ModTime,
		RefCount:     oi.RefCount,
		isNewReal:    oi.isNewReal,
		isDirty:      oi.isDirty,
		isDeleted:    oi.isDeleted,
		isProcessing: oi.isProcessing,
	}
}

func (oi *ObjectInfo) String() string {
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

func (oi *ObjectInfo) Bytes() []byte {
	if debug {
		if oi.ID.IsZero() {
			panic("should not happen")
		}
		if oi.Hash.IsZero() {
			panic("should not happen")
		}
		if oi.OwnerID.IsZero() {
			panic("should not happen")
		}
	}
	bz := make([]byte, 0, 100)
	bz = append(bz, sizedBytes(oi.ID.Bytes())...)
	bz = append(bz, sizedBytes(oi.Hash.Bytes())...)
	bz = append(bz, sizedBytes(oi.OwnerID.Bytes())...)
	bz = append(bz, varintBytes(int64(oi.ModTime))...)
	bz = append(bz, varintBytes(int64(oi.RefCount))...)
	return bz
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
	oi.OwnerID = po.GetObjectID()
	oi.owner = po
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

func (oi *ObjectInfo) GetIsNewReal() bool {
	return oi.isNewReal
}

func (oi *ObjectInfo) SetIsNewReal(x bool) {
	oi.isNewReal = x
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

func (oi *ObjectInfo) GetIsDeleted() bool {
	return oi.isDeleted
}

func (oi *ObjectInfo) SetIsDeleted(x bool, mt uint64) {
	// NOTE: Don't over-write modtime.
	// Consider adding a DelTime, or just log it somewhere, or
	// continue to ignore it.
	oi.isDirty = x
}

func (oi *ObjectInfo) GetIsProcessing() bool {
	return oi.isProcessing
}

func (oi *ObjectInfo) SetIsProcessing(x bool) {
	oi.isProcessing = x
}

func (oi *ObjectInfo) GetIsTransient() bool {
	return false
}

func (tv *TypedValue) GetFirstObject(store Store) Object {
	switch cv := tv.V.(type) {
	case PointerValue:
		// TODO: in the future, consider skipping the base if persisted
		// ref-count would be 1, e.g. only this pointer refers to
		// something in it; in that case, ignore the base.  That will
		// likely require maybe a preperation step in persistence
		// ( or unlikely, a second type of ref-counting).
		if cv.Base != nil {
			return cv.Base.(Object)
		} else {
			return cv.TV.GetFirstObject(store)
		}
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
	case nativeValue:
		panic("realm logic for native values not supported")
	case *Block:
		return cv
	default:
		return nil
	}
}

//----------------------------------------
// ExtendedObject
// ExtendedObject is for storing native arrays, slices, structs, maps, as
// well as Gno maps. It implements Object for gno maps, but not for native
// types which are not supported by realm persistence.  ExtendedObject is
// required for *MapValue for the machine state to be persistable between
// slot access and assignment to said slot.

type ExtendedObject struct {
	BaseMap    *MapValue    // if base is gno map
	BaseNative *nativeValue // if base is native array/slice/struct/map.
	Index      TypedValue   // integer index or arbitrary map key
	Path       ValuePath    // value path for (native) selectors
}

// implements Value
func (eo ExtendedObject) assertValue() {
}

// implements Value
func (eo ExtendedObject) String() string {
	if eo.BaseMap != nil {
		return eo.BaseMap.String()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetObjectInfo() *ObjectInfo {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetObjectInfo()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetObjectID() ObjectID {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetObjectID()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) MustGetObjectID() ObjectID {
	if eo.BaseMap != nil {
		return eo.BaseMap.MustGetObjectID()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetObjectID(oid ObjectID) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetObjectID(oid)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetHash() ValueHash {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetHash()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetHash(vh ValueHash) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetHash(vh)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetOwner() Object {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetOwner()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetOwnerID() ObjectID {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetOwnerID()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetOwner(obj Object) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetOwner(obj)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsOwned() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsOwned()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsReal() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsReal()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetModTime() uint64 {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetModTime()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) IncRefCount() int {
	if eo.BaseMap != nil {
		return eo.BaseMap.IncRefCount()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) DecRefCount() int {
	if eo.BaseMap != nil {
		return eo.BaseMap.DecRefCount()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetRefCount() int {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetRefCount()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsNewReal() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsNewReal()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetIsNewReal(b bool) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetIsNewReal(b)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsDirty() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsDirty()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetIsDirty(b bool, mt uint64) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetIsDirty(b, mt)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsDeleted() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsDeleted()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetIsDeleted(b bool, mt uint64) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetIsDeleted(b, mt)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsProcessing() bool {
	if eo.BaseMap != nil {
		return eo.BaseMap.GetIsProcessing()
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) SetIsProcessing(b bool) {
	if eo.BaseMap != nil {
		eo.BaseMap.SetIsProcessing(b)
	} else {
		panic("native values are not realm compatible")
	}
}

func (eo ExtendedObject) GetIsTransient() bool {
	if eo.BaseMap != nil {
		return false
	} else {
		return true // native values cannot be realm persisted.
	}
}

/*
func (eo ExtendedObject) ValueImage(rlm *Realm, owned bool) *ValueImage {
	if eo.BaseMap != nil {
		return eo.BaseMap.ValueImage(rlm, owned)
	} else {
		panic("native values are not realm compatible")
	}
}
*/
