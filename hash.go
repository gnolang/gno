package gno

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const HashSize = 20

type Hashlet [HashSize]byte

func (h Hashlet) IsZero() bool {
	return h == Hashlet{}
}

func HashBytes(bz []byte) (res Hashlet) {
	hash := sha256.Sum256(bz)
	copy(res[:], hash[:HashSize])
	return
}

func leafHash(bz []byte) (res Hashlet) {
	buf := make([]byte, 1+len(bz))
	buf[0] = 0x00
	copy(buf[1:], bz)
	return HashBytes(buf)
}

func innerHash(h1, h2 Hashlet) (res Hashlet) {
	buf := make([]byte, 1+HashSize*2)
	buf[0] = 0x01
	copy(buf[1:1+HashSize], h1[:])
	copy(buf[1+HashSize:], h2[:])
	return HashBytes(buf)
}

//----------------------------------------
// ValueHash
//
// The ValueHash of a typed value is a unique deterministic
// accountable fingerprint of that typed value, and can be used
// to prove the value or any part of its value which is
// accessible from the Gno language.
//
// For example, the ValueHash of a primitive value is simply
// the "leaf hash" of its TypedValuePreimage. The ValueHash of
// a non-primitive object is the merkle root hash of its
// elements' TypedElemPreimage's.  The ValueHash of a nil
// interface value is zero.
//
// `ValueHash := lh(TypedValuePreimage)`
// `ValueHash := zero` if nil interface.
//
// `TypedValuePreimage := sz(TypeID),ValuePreimage`
//  * TypeID is a byte-slice if byte-array.
//
// `ValuePreimage := 0x00` if typed-nil.
// `ValuePreimage := 0x01,sz(pb(.))` if primitive.
// `ValuePreimage := 0x02,sz(vh(*ptr))` if ptr.
// `ValuePreimage := 0x03,sz(data)` if byte-array.
// `ValuePreimage := 0x04,sz(ElemsHash)` if non-nil object.
// `ValuePreimage := 0x05,sz(vh(base)),off,len,max if slice.
//
// `ElemsHash := lh(TypedElemPreimage)` if object w/ 1 elem.
// `ElemsHash := ih(rh(Left),rh(Right))` if object w/ 2+ elems.
//
// `TypedElemPreimage := sz(TypeID),ElemPreimage`
// `TypedElemPreimage := nil` if nil interface.
//
// `ElemPreimage := ...` (of each array/struct/block element)
// `ElemPreimage := 0x10` if typed-nil.
// `ElemPreimage := 0x11,sz(ObjectID)` if borrowed.
// `ElemPreimage := 0x12,sz(ObjectID),sz(vh(.))` if owned.
// `ElemPreimage := 0x13,sz(nil),sz(vh(.))` prim/ptr/slice.
//  * ownership passed through for pointers/slices/arrays.
//
// * sz() means (uvarint) size-prefixed bytes.
// * vh() means .ValueHash().
// * rh() are inner ElemsHashs.
// * lh() means leafHash(x) := hash(0x00,x)
// * ih() means innerHash(x,y) := hash(0x01,x,y)
// * pb() means .PrimitiveBytes().
// * off,len,max and other integers are uvarint encoded.
// * len(Left) is always 2^x, x=0,1,2,...
// * Right may be zero (if len(Left+Right) not 2^x)
//
// If a pointer value is owned (e.g. field tagged "owned"), the
// pointer's base if present must not already be owned.  If a
// pointer value is not owned, but refers to a value that has a
// refcount of 1, it is called "run-time" owned, and the value
// bytes include the hash of the referred value or object as if
// owned; the value bytes also include the object-id of the
// "run-time" owned object as if it were persisted separately
// from its base object, but implementations may choose to
// inline the serialization of "run-time" owned objects anyway.
//
// If an object is owned, the value hash of elements is
// included, otherwise, the value hash of elements is not
// included except for objects with refcount=1.  If owned but
// any of the elements are already owned, or if not owned but
// any of the elements have refcount>1, preimage derivation
// panics.

type ValueHash Hashlet

//----------------------------------------
// TypedValuePreimage

type TypedValuePreimage struct {
	TypeID         // never nil
	ValType        // 0:nil,1:prim,2:ptr,3:data,4:obj,5:slice
	Data    []byte // if ValType=prim,ptr,data,obj,slice
	Offset  int    // if ValType=slice
	Length  int    // if ValType=slice
	Maxcap  int    // if ValType=slice

	preimages []TypedElemPreimage // for debugging
}

type ValType byte

const (
	ValTypeNil       = ValType(0x00)
	ValTypePrimitive = ValType(0x01)
	ValTypePointer   = ValType(0x02)
	ValTypeData      = ValType(0x03)
	ValTypeObject    = ValType(0x04)
	ValTypeSlice     = ValType(0x05)
)

func (tvp *TypedValuePreimage) Bytes() []byte {
	if debug {
		if tvp.TypeID.IsZero() {
			panic("should not happen")
		}
	}
	buf := sizedBytes(tvp.TypeID[:])
	buf = append(buf, byte(tvp.ValType))
	switch tvp.ValType {
	case ValTypeNil:
		return buf
	case ValTypePrimitive:
		fallthrough
	case ValTypePointer:
		fallthrough
	case ValTypeData:
		fallthrough
	case ValTypeObject:
		buf = append(buf, sizedBytes(tvp.Data)...)
		return buf
	case ValTypeSlice:
		buf = append(buf, sizedBytes(tvp.Data)...)
		buf = append(buf, uvarintBytes(uint64(tvp.Offset))...)
		buf = append(buf, uvarintBytes(uint64(tvp.Length))...)
		buf = append(buf, uvarintBytes(uint64(tvp.Maxcap))...)
		return buf
	default:
		panic("should not happen")
	}
}

func (tvp *TypedValuePreimage) ValueHash() ValueHash {
	return ValueHash(leafHash(tvp.Bytes()))
}

//----------------------------------------
// ElemPreimage

type TypedElemPreimage struct {
	TypeID    // never nil
	ElemType  // 0x10:typed-nil,0x11:brwd,0x12:owned,0x13:other
	ObjectID  // if ElemType=borrowed,owned
	ValueHash // if ElemType=other (valuehash)
}

type ElemType byte

const (
	ElemTypeTypedNil = ElemType(0x10)
	ElemTypeBorrowed = ElemType(0x11)
	ElemTypeOwned    = ElemType(0x12)
	ElemTypeOther    = ElemType(0x13)
)

func (tep *TypedElemPreimage) Bytes() []byte {
	if debug {
		if tep.TypeID.IsZero() {
			panic("should not happen")
		}
	}
	buf := sizedBytes(tep.TypeID[:])
	buf = append(buf, byte(tep.ElemType))
	switch tep.ElemType {
	case ElemTypeTypedNil:
		if debug {
			if !tep.ObjectID.IsZero() {
				panic("should not happen")
			}
			if tep.ValueHash != (ValueHash{}) {
				panic("should not happen")
			}
		}
		return buf
	case ElemTypeBorrowed:
		if debug {
			if tep.ObjectID.IsZero() {
				panic("should not happen")
			}
			if tep.ValueHash != (ValueHash{}) {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(tep.ObjectID.Bytes())...)
		return buf
	case ElemTypeOwned:
		if debug {
			if tep.ObjectID.IsZero() {
				panic("should not happen")
			}
			if tep.ValueHash == (ValueHash{}) {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(tep.ObjectID.Bytes())...)
		buf = append(buf, sizedBytes(tep.ValueHash[:])...)
		return buf
	case ElemTypeOther:
		if debug {
			if !tep.ObjectID.IsZero() {
				panic("should not happen")
			}
			if tep.ValueHash == (ValueHash{}) {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(tep.ValueHash[:])...)
		return buf
	default:
		panic("should not happen")
	}
}

func (tep *TypedElemPreimage) LeafHash() Hashlet {
	return leafHash(tep.Bytes())
}

func ElemsHashFromElements(tepz []TypedElemPreimage) Hashlet {
	// special case if nil
	if tepz == nil {
		return Hashlet{}
	}
	// translate to leaf hashes
	hz := make([]Hashlet, ((len(tepz)+1)/2)*2)
	for i, tvp := range tepz {
		hz[i] = tvp.LeafHash()
	}
	// merkle-ize
	for 1 < len(hz) {
		// pair-wize hash
		for i := 0; i < len(hz); i += 2 {
			h1 := hz[i]
			h2 := hz[i+1]
			hz[i/2] = innerHash(h1, h2)
		}
		// resize hz to half (but even)
		hz = hz[:len(hz)/2]
		if len(hz)%2 == 1 {
			hz = append(hz, Hashlet{})
		}
	}
	return hz[0]
}

//----------------------------------------
// Value.TypedValuePreimage
// Value.TypedElemPreimages

// NOTE: needed for SliceValue.TypedValuePreimage.  A
// SliceValue shouldn't care about the underlying array type;
// and we may improve the preimage standard for slices that own
// their data, where the base length matters even less.
//
// If the array is a bytearray, the ValueHash is computed as if
// it were a primitive value. In that case the result's
// preimage has TypeID set to lt.TypeID().  If this *ArrayValue
// is referenced in a *SliceValue, the preimage for the
// *SliceValue includes the HashValue of the base *ArrayValue;
// and in that case baseOf(lt) is a *SliceType.  It follos that
// lt may be an *ArrayType, a *SliceType, or a *DeclaredType
// whose base type is an array or slice type.
func (av *ArrayValue) TypedValuePreimage(
	rlm *Realm, owned bool, lt Type) TypedValuePreimage {

	avl := av.GetLength()
	et := lt.Elem()
	if et.Kind() == Uint8Kind {
		// `ValuePreimage := 0x03,sz(data)` if byte-array.
		return av.TypedValuePreimage(rlm, owned, lt)
		data := av.Data
		if data == nil {
			data = make([]byte, avl)
			copyListToData(
				data[:avl],
				av.List[:avl],
			)
		}
		return TypedValuePreimage{
			TypeID:  lt.TypeID(),
			ValType: ValTypeData,
			Data:    data,
		}
	}
	// `ValuePreimage := 0x04,sz(ElemsHash)` if non-nil object.
	tepz := av.TypedElemPreimages(rlm, owned, lt)
	eh := ElemsHashFromElements(tepz)
	return TypedValuePreimage{
		TypeID:    lt.TypeID(),
		ValType:   ValTypeObject,
		Data:      eh[:],
		preimages: tepz,
	}
}

// NOTE: lt may be array or slice type.
func (av *ArrayValue) TypedElemPreimages(
	rlm *Realm, owned bool, lt Type) []TypedElemPreimage {

	avl := av.GetLength()
	if avl == 0 {
		return nil
	}
	et := lt.Elem()
	// Sanity check.
	if et.Kind() == Uint8Kind {
		panic("ArrayValue of bytes has no TypedElemPreimages," +
			"call TypedValuePreimage() instead")
	}
	// General (list) case.
	tepz := make([]TypedElemPreimage, avl)
	for i := 0; i < avl; i++ {
		ev := &av.List[i]
		if ev.IsUndefined() && et != nil {
			ev.T = et // set in place.
		}
		tepz[i] = ev.TypedElemPreimage(rlm, owned)
	}
	return tepz
}

func (sv *SliceValue) TypedValuePreimage(
	rlm *Realm, owned bool, st Type) TypedValuePreimage {

	if sv.Base == nil {
		return TypedValuePreimage{} // nil slice
	}
	// If (self, base) is:
	//  - (owned, already-owned):
	//    * panic (ownership conflict)
	//  - (owned, not-owned):
	//    * TODO: trim array to slice window first
	//    * save & hash array w/ its object_id
	//    * leafHash(Base.TVP(true),Offset,Length,Maxcap)
	//  - (not-owned, already-owned):
	//    * leafHash(Base.TVP(false),Offset,Length,Maxcap)
	//  - (not-owned, not-owned & refcount=1):
	//    * continue as (runtime) owned
	//  - (not-owned, not-owned & refcount=2+):
	//    * panic (ambiguous)
	var baseAsOwned bool
	if owned {
		if sv.Base.GetIsOwned() {
			panic("ownership conflict")
		} else {
			baseAsOwned = true
		}
	} else {
		if sv.Base.GetIsOwned() {
			baseAsOwned = false
		} else {
			if debug {
				if sv.Base.GetRefCount() == 0 {
					panic("should not happen")
				}
			}
			if sv.Base.GetRefCount() == 1 {
				baseAsOwned = true
			} else {
				panic("ambiguous ownership")
			}
		}
	}
	// `ValuePreimage := 0x05,sz(vh(base)),off,len,max if slice.
	tvp := sv.Base.TypedValuePreimage(rlm, baseAsOwned, st)
	bvh := tvp.ValueHash()
	return TypedValuePreimage{
		TypeID:  st.TypeID(),
		ValType: ValTypeSlice,
		Data:    bvh[:],
		Offset:  sv.Offset,
		Length:  sv.Length,
		Maxcap:  sv.Maxcap,
	}
}

func (sv *StructValue) TypedElemPreimages(
	rlm *Realm, owned bool, st Type) []TypedElemPreimage {

	nf := len(sv.Fields)
	if nf == 0 {
		return nil
	}
	tvpz := make([]TypedElemPreimage, nf)
	for i := 0; i < nf; i++ {
		fv := &sv.Fields[i] // ref
		ft := baseOf(st).(*StructType).Fields[i]
		if fv.IsUndefined() && ft.Type.Kind() != InterfaceKind {
			fv.T = ft.Type // mutates sv
		}
		tvpz[i] = fv.TypedElemPreimage(rlm, owned)
	}
	return tvpz
}

// TODO split into keyOwned and valueOwned with new tags.
func (mv *MapValue) TypedElemPreimages(
	rlm *Realm, owned bool, mt Type) []TypedElemPreimage {

	if mv.vmap == nil {
		return nil
	}

	ms := mv.List.Size
	tvpz := make([]TypedElemPreimage, ms*2) // even:key odd:val
	head := mv.List.Head
	mkt := mt.(*MapType).Key
	mvt := mt.(*MapType).Value
	kii := mkt.Kind() == InterfaceKind
	vii := mvt.Kind() == InterfaceKind
	// create deterministic list from mv.List
	for i := 0; i < ms; i++ {
		key, val := &head.Key, &head.Value // ref
		if key.IsUndefined() && !kii {
			key.T = mkt // mutates mv
		}
		if val.IsUndefined() && !vii {
			val.T = mvt // mutates mv
		}
		tvpz[i*2+0] = key.TypedElemPreimage(rlm, owned)
		tvpz[i*2+1] = val.TypedElemPreimage(rlm, owned)
		// done
		head = head.Next
		if debug {
			if i == ms-1 {
				if head != nil {
					panic("should not happen")
				}
			}
		}
	}
	return tvpz
}

func (b *Block) ValueHash(
	rlm *Realm, owned bool, pt Type) ValueHash {

	// Block values are not yet implemented, and we don't yet
	// support the persistence of closures or goroutine
	// snapshots.  TODO: Implement this, and support the
	// persistence of closures, or break out into more TODOs.
	panic("not yet implemented")
}

//----------------------------------------
// *TypedValue.ValueHash

func (tv *TypedValue) ValueHash(rlm *Realm, owned bool) ValueHash {
	if debug {
		if tv.T.Kind() == InterfaceKind {
			panic("should not happen")
		}
	}
	if tv.IsUndefined() {
		// `ValueHash := zero` if nil interface.
		return ValueHash{}
	} else {
		tvp := tv.TypedValuePreimage(rlm, owned)
		// `ValueHash := lh(TypedValuePreimage)`
		return tvp.ValueHash()
	}
}

func (tv *TypedValue) TypedValuePreimage(rlm *Realm, owned bool) TypedValuePreimage {
	if tv.IsUndefined() {
		panic("undefined value has no TypedValuePreimage")
	}
	tid := tv.T.TypeID()
	if tv.V == nil { // primitive or nil
		if _, ok := baseOf(tv.T).(PrimitiveType); ok {
			// `ValuePreimage := 0x01,sz(pb(.))` if primitive.
			pbz := tv.PrimitiveBytes()
			return TypedValuePreimage{
				TypeID:  tid,
				ValType: ValTypePrimitive,
				Data:    pbz,
			}
		} else {
			// `ValuePreimage := 0x00` if typed-nil.
			return TypedValuePreimage{
				TypeID:  tid,
				ValType: ValTypeNil,
			}
		}
	} else { // non-nil object.
		switch baseOf(tv.T).(type) {
		case PointerType:
			pv := tv.V.(PointerValue)
			if pv.TypedValue == nil {
				// `ValuePreimage := 0x02,sz(vh(*ptr))` if ptr.
				// `ValueHash := zero` if nil.
				return TypedValuePreimage{
					TypeID:  tid,
					ValType: ValTypePointer,
					Data:    nil,
				}
			} else {
				// `ValuePreimage := 0x02,sz(vh(*ptr))` if ptr.
				vh := tv.ValueHash(rlm, owned)
				return TypedValuePreimage{
					TypeID:  tid,
					ValType: ValTypePointer,
					Data:    vh[:],
				}
			}
		case *ArrayType:
			av := tv.V.(*ArrayValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			return av.TypedValuePreimage(rlm, owned, tv.T)
		case *SliceType:
			sv := tv.V.(*SliceValue)
			// `ValuePreimage :=
			//    0x05,sz(vh(base)),off,len,max if slice.
			//  * TypeID is a byte-slice if byte-array.  NOTE:
			//  Slices do not have access to an underlying array's
			//  *DeclaredType if any.
			return sv.TypedValuePreimage(rlm, owned, tv.T)
		case *StructType:
			sv := tv.V.(*StructValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			tepz := sv.TypedElemPreimages(rlm, owned, tv.T)
			eh := ElemsHashFromElements(tepz)
			return TypedValuePreimage{
				TypeID:    tid,
				ValType:   ValTypeObject,
				Data:      eh[:],
				preimages: tepz,
			}
		case *MapType:
			mv := tv.V.(*MapValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			tepz := mv.TypedElemPreimages(rlm, owned, tv.T)
			eh := ElemsHashFromElements(tepz)
			return TypedValuePreimage{
				TypeID:    tid,
				ValType:   ValTypeObject,
				Data:      eh[:],
				preimages: tepz,
			}
		default:
			panic(fmt.Sprintf(
				"unexpected type for TypedValuePreimage(): %s",
				tv.T.String()))
		}
	}
}

func (tv *TypedValue) TypedElemPreimage(rlm *Realm, owned bool) TypedElemPreimage {
	if tv.IsUndefined() {
		// `TypedElemPreimage := nil` if nil interface.
		return TypedElemPreimage{} // nil
	} else if tv.T.Kind() == InterfaceKind {
		if debug {
			panic("should not happen")
		}
	}
	tid := tv.T.TypeID()
	if tv.V == nil {
		if _, ok := baseOf(tv.T).(PrimitiveType); ok {
			// `ElemPreimage := 0x13,sz(nil),sz(vh(.))` prim/ptr/slice.
			// `ValueHash := lh(TypedValuePreimage)` ...
			tvp := tv.TypedValuePreimage(rlm, owned)
			vh := tvp.ValueHash()
			return TypedElemPreimage{
				TypeID:    tid,
				ElemType:  ElemTypeOther,
				ValueHash: vh,
			}
		} else {
			// `ElemPreimage := 0x10` if typed-nil.
			return TypedElemPreimage{
				TypeID:   tid,
				ElemType: ElemTypeTypedNil,
			}
		}
	} else {
		switch baseOf(tv.T).(type) {
		case PointerType, *SliceType:
			// `ElemPreimage := 0x13,sz(nil),sz(vh(.))`
			// 	 if prim/ptr/slice.
			// `ValueHash := lh(TypedValuePreimage)`
			tvp := tv.TypedValuePreimage(rlm, owned)
			vh := tvp.ValueHash()
			return TypedElemPreimage{
				TypeID:    tid,
				ElemType:  ElemTypeOther,
				ValueHash: vh,
			}
		case *ArrayType, *StructType, *MapType:
			var obj Object = tv.V.(Object)
			if !owned {
				rc := obj.GetRefCount()
				if debug {
					if rc <= 0 {
						panic("should not happen")
					}
				}
				if rc == 1 {
					owned = true
				}
			}
			oid := obj.GetObjectID()
			if owned {
				// `ElemPreimage := 0x12,sz(ObjectID),sz(vh(.))` if owned.
				tvp := tv.TypedValuePreimage(rlm, owned)
				vh := tvp.ValueHash()
				return TypedElemPreimage{
					TypeID:    tid,
					ElemType:  ElemTypeOwned,
					ObjectID:  oid,
					ValueHash: vh,
				}
			} else {
				// `ElemPreimage := 0x11,sz(ObjectID)` if borrowed.
				return TypedElemPreimage{
					TypeID:   tid,
					ElemType: ElemTypeBorrowed,
					ObjectID: oid,
				}
			}
		default:
			panic(fmt.Sprintf(
				"unexpected type for elem preimaging: %s",
				tv.T.String()))
		}
	}
}

//----------------------------------------
// misc

func uvarintBytes(u uint64) []byte {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], u)
	return buf[0:n]
}

func sizedBytes(bz []byte) []byte {
	bz2 := make([]byte, len(bz)+10)
	n := binary.PutUvarint(bz2[:10], uint64(len(bz)))
	copy(bz2[n:n+len(bz)], bz)
	return bz2
}
