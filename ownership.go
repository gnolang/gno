package gno

import (
	"encoding/binary"
	"fmt"
)

type ObjectID struct {
	RealmID        // base
	Ordinal uint64 // counter
}

func (oid ObjectID) String() string {
	if oid.RealmID.IsZero() {
		// XXX what's at the very top?
		return fmt.Sprintf("OIDNONE:%d", oid.Ordinal)
	} else {
		return fmt.Sprintf("OID%X:%d",
			oid.RealmID.Bytes(), oid.Ordinal)
	}
}

func (oid ObjectID) Bytes() []byte {
	bz := make([]byte, HashSize+8)
	copy(bz[:HashSize], oid.RealmID.Bytes())
	binary.BigEndian.PutUint64(
		bz[HashSize:], uint64(oid.Ordinal))
	return bz
}

// TODO: make faster by making RealmID a pointer
// and enforcing that the value of RealmID is never zero.
func (oid ObjectID) IsZero() bool {
	if debug {
		if oid.RealmID.IsZero() && oid.Ordinal != 0 {
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
	GetOwner() Object
	GetOwnerID() ObjectID
	SetOwner(Object)
	GetIsOwned() bool
	GetIsReal() bool
	IncRefCount() int
	DecRefCount() int
	GetRefCount() int
	GetIsNewReal() bool
	SetIsNewReal(bool)
	GetIsDirty() bool
	SetIsDirty(bool)
	GetIsDeleted() bool
	SetIsDeleted(bool)

	// Saves to realm along the way if owned, and also (dirty
	// or new).
	ValueImage(rlm *Realm, owned bool) *ValueImage
	ElemImages(rlm *Realm, owned bool) []ElemImage
}

var _ Object = &ArrayValue{}
var _ Object = &StructValue{}
var _ Object = &MapValue{}
var _ Object = &Block{}

type ObjectInfo struct {
	ID        ObjectID  // set if real.
	Hash      ValueHash // zero if dirty.
	OwnerID   ObjectID  // parent in the ownership tree.
	RefCount  int       // deleted/gc'd if 0.
	isNewReal bool
	isDirty   bool
	isDeleted bool

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

func (oi *ObjectInfo) GetValueHash() ValueHash {
	return oi.Hash
}

func (oi *ObjectInfo) SetValueHash(vh ValueHash) {
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

func (oi *ObjectInfo) IncRefCount() int {
	oi.RefCount++
	return oi.RefCount
}

func (oi *ObjectInfo) DecRefCount() int {
	oi.RefCount--
	if debug {
		if oi.RefCount < 0 {
			panic("should not happen")
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

func (oi *ObjectInfo) SetIsDirty(x bool) {
	if x {
		oi.Hash = ValueHash{}
	}
	oi.isDirty = x
}

func (oi *ObjectInfo) GetIsDeleted() bool {
	return oi.isDeleted
}

func (oi *ObjectInfo) SetIsDeleted(x bool) {
	oi.isDirty = x
}

// Returns the value as an object if it is an object,
// or is a pointer or slice of an object.
func (tv *TypedValue) GetObject() Object {
	switch cv := tv.V.(type) {
	case PointerValue:
		return cv.Base
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
		rov, ok := cv.Receiver.(Object)
		if ok {
			return rov
		} else {
			return nil
		}
	case nativeValue:
		panic("native not compatible with realm logic")
	case blockValue:
		if cv.Block == nil {
			panic("should not happen")
		}
		return cv.Block
	default:
		return nil
	}
}
