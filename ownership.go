package gno

import (
	"encoding/binary"
	"fmt"
)

type ObjectID struct {
	RealmID        // base
	NewTime uint64 // time created
}

func (oid ObjectID) String() string {
	if oid.RealmID.IsZero() {
		// XXX what's at the very top?
		return fmt.Sprintf("OIDNONE:%d", oid.NewTime)
	} else {
		return fmt.Sprintf("OID%X:%d",
			oid.RealmID.Bytes(), oid.NewTime)
	}
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
var _ Object = &MapValue{}
var _ Object = &Block{}

type ObjectInfo struct {
	ID           ObjectID  // set if real.
	Hash         ValueHash // zero if dirty.
	OwnerID      ObjectID  // parent in the ownership tree.
	ModTime      uint64    // time last updated.
	RefCount     int       // deleted/gc'd if 0.
	isNewReal    bool
	isDirty      bool
	isDeleted    bool
	isProcessing bool

	owner Object // mem reference to owner.
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

// Returns the value as an object if it is an object,
// or is a pointer or slice of an object.
func (tv *TypedValue) GetObject() Object {
	switch cv := tv.V.(type) {
	case PointerValue:
		// TODO: In terms of defining the object dependency graph,
		// whether the relevant object is the pointer base or
		// the pointed object (.base or .typedvalue) depends
		// on the number of references to the base. In the future
		// when supporting ref-counted and weak references,
		// calculate this on the fly or with a pre-pass.
		return cv.TypedValue.GetObject()
	case *ArrayValue:
		return cv
	case *SliceValue:
		if cv.Base == nil {
			// otherwise `return cv.Base` returns a typed-nil.
			return nil
		} else {
			return cv.Base
		}
	case *StructValue:
		return cv
	case *FuncValue:
		return nil
	case *MapValue:
		return cv
	case BoundMethodValue:
		rov, ok := cv.Receiver.V.(Object)
		if ok {
			return rov
		} else {
			return nil
		}
	case nativeValue:
		// native values don't work with realms,
		// but this function shouldn't happen.
		// XXX panic?
		return nil
	case blockValue:
		if cv.Block == nil {
			panic("should not happen")
		}
		return cv.Block
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
