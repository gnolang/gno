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

func (oid ObjectID) IsPackageID() bool {
	// all package objects have newtime 1.
	return !oid.PkgID.IsZero() && oid.NewTime == 1
}

// IsZero returns true iff the ObjectID is completely empty (both
// PkgID and NewTime are zero). This is the "totally empty" state
// — used to detect "no owner exists" / "transient/never-stamped".
//
// An ObjectID has three states:
//
//	empty:     PkgID zero, NewTime zero        (never went through allocator)
//	allocated: PkgID set,  NewTime zero        (set at construction by the allocator)
//	finalized: PkgID set,  NewTime ≥ 1         (real, persisted)
//
// Use IsFinalized() for "has a real persisted identity" and
// GetIsReal() (on ObjectInfo) as the convenience equivalent.
func (oid ObjectID) IsZero() bool {
	if debug {
		// The impossible state is PkgID zero + NewTime non-zero.
		// PkgID set + NewTime zero is the allocated-but-unfinalized
		// state.
		if oid.PkgID.IsZero() && oid.NewTime != 0 {
			panic("invariant: NewTime set but PkgID zero")
		}
	}
	return oid.PkgID.IsZero() && oid.NewTime == 0
}

// IsFinalized returns true iff the ObjectID has been stamped with a
// NewTime by assignNewObjectID — i.e., it has a real persisted
// identity. The allocated-but-unfinalized state (PkgID set, NewTime
// zero) returns false.
func (oid ObjectID) IsFinalized() bool {
	return oid.NewTime != 0
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
	SetNewTime(t uint64) // partial stamp during assignNewObjectID
	SetPkgID(p PkgID)    // partial stamp at allocation time
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
	SetIsDeleted(bool)
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
	_ Object = &PackageValue{}
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
	if !oi.ID.IsFinalized() {
		panic("unexpected non-finalized object id")
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
	return oi.OwnerID
}

func (oi *ObjectInfo) GetIsOwned() bool {
	return !oi.OwnerID.IsZero()
}

// GetIsReal returns true iff the object has a finalized ObjectID
// (NewTime ≥ 1). Allocated-but-unfinalized objects (PkgID set,
// NewTime zero) return false. Note: does not return true for
// new-reals (those waiting for assignNewObjectID at finalize time).
func (oi *ObjectInfo) GetIsReal() bool {
	return oi.ID.IsFinalized()
}

// SetNewTime stamps only the NewTime portion of the ObjectID,
// preserving any pre-existing PkgID set at allocation time.
// Used by assignNewObjectID.
func (oi *ObjectInfo) SetNewTime(t uint64) {
	oi.ID.NewTime = t
}

// SetPkgID stamps only the PkgID portion of the ObjectID,
// preserving any pre-existing NewTime. Used by allocator
// constructors to stamp authority at allocation.
func (oi *ObjectInfo) SetPkgID(p PkgID) {
	oi.ID.PkgID = p
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

// SetIsDeleted marks the object as deleted. The deletion is just a
// tombstone marker, not a clock — under cross-realm finalize this
// avoids a "myrealm's clock stamps yourrealm's tombstone" semantic
// discrepancy.
func (oi *ObjectInfo) SetIsDeleted(x bool) {
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
		return cv
	case *Block:
		return cv
	case RefValue:
		if cv.PkgPath != "" {
			// Constructed by preprocessor from package name exprs
			// (or derived implicitly for local package names).
			// These may refer to package values not yet
			// real/persisted; this function should not handle it.
			panic("GetFirstObject() cannot handle RefValue{PkgPath}")
		}
		oo := store.GetObject(cv.ObjectID)
		tv.V = oo
		return oo
	case *HeapItemValue:
		// should only appear in PointerValue.Base or
		// closure capture; if you need to implement
		// this, probably doing it wrong.
		panic("invalid usage of GetFirstObject() on heap item")
	default:
		return nil
	}
}

// IsReadonlyBy returns true if tv is a real object owned by a realm
// other than rid (i.e., residing in an external realm).
//
// ownPkgID is the executing package's PkgID (m.Package.PkgID). An object
// stamped with it is the executing package's own package-level data, which
// the package may always read/copy regardless of rid — e.g. stdlib or a /p/
// library reading its own immutable tables while running under a caller's
// realm (those callables don't borrow, so m.Realm is the caller's, not the
// library's). Pass a zero PkgID to disable this exemption.
//
// This is different from GetFirstObject in two significant ways:
//  1. IsReadonlyBy does not go through RefValues; for this reason, it
//     also doesn't need a store to fetch the nested object.
//  2. If a pointer's HeapItemValue is unreal, only the object id of
//     its underlying Value is considered.
//  3. If a pointer's HeapItemValue is real, both the object id of
//     the heap item value AND its internal value is considered.
//
// This function controls heavily the behaviour of
// [Machine.IsReadonly], and thus cross-realm write authority.
func (tv *TypedValue) IsReadonlyBy(rid, ownPkgID PkgID) bool {
	var tvoid ObjectID
	switch cv := tv.V.(type) {
	case PointerValue:
		if cv.Base == nil {
			return false // free floating
		}
		if hiv, ok := cv.Base.(*HeapItemValue); ok {
			// Also need to check the heap item value.
			// NOTE: It is possible for the value to be
			// external while the heap item itself is
			// not.
			// See test/files/zrealm_crossrealm25a.gno.
			if hiv.Value.IsReadonlyBy(rid, ownPkgID) {
				return true
			}
			// An unreal HIV is a transient heap-promotion wrapper
			// for an escaping local (closure capture, new(T), &T{},
			// etc.) — not a realm-owned slot. The alloc-site PkgID
			// stamp is incidental, so skip the PkgID gate while
			// the HIV is unreal. Once persisted (NewTime>0) the
			// standard gate applies: cross-realm writes to a
			// persisted captured slot are rejected so callers must
			// use an explicit crossing function in the owning
			// realm, keeping all cross-realm side effects
			// syntactically visible via `cross`.
			if !hiv.GetIsReal() {
				return false
			}
			tvoid = hiv.GetObjectID()
		} else {
			tvoid = cv.Base.(ObjectIDer).GetObjectID()
		}
	case *ArrayValue:
		tvoid = cv.GetObjectID()
	case *SliceValue:
		tvoid = cv.Base.(ObjectIDer).GetObjectID()
	case *StructValue:
		tvoid = cv.GetObjectID()
	case *FuncValue:
		tvoid = cv.GetObjectID()
	case *MapValue:
		tvoid = cv.GetObjectID()
	case *BoundMethodValue:
		tvoid = cv.GetObjectID()
	case *PackageValue:
		tvoid = cv.GetObjectID()
	case *Block:
		tvoid = cv.GetObjectID()
	case RefValue:
		if cv.PkgPath != "" {
			// Constructed by preprocessor from package name exprs
			// (or derived implicitly for local package names).
			// These may refer to package values not yet
			// real/persisted; this function should not handle it.
			// It is should be handled by Machine.IsReadonly().
			panic("IsReadonlyBy() cannot handle RefValue{PkgPath}")
		}
		tvoid = cv.GetObjectID()
	case *HeapItemValue:
		// should only appear in PointerValue.Base or
		// closure capture; if you need to implement
		// this, probably doing it wrong.
		panic("invalid usage of IsReadonly() on heap item")
	default:
		// tv is not an object ("first object" ID is zero)
		return false // e.g. primitive
	}
	// tv is an unreal object (no object id)
	if tvoid.IsZero() {
		return false
	}
	// tv is an object residing in external realm — unless it is the
	// executing package's own package-level data (stamped ownPkgID),
	// which the package may always read/copy regardless of m.Realm.
	if !tvoid.PkgID.eq(rid) && !tvoid.PkgID.eq(ownPkgID) {
		return true
	}
	return false
}
