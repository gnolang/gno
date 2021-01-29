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
// `ValuePreimage := 0x05,sz(vp(base)),off,len,max if slice.
//
// `ElemsHash := lh(TypedElemPreimage)` if object w/ 1 elem.
// `ElemsHash := ih(eh(Left),eh(Right))` if object w/ 2+ elems.
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
// * eh() are inner ElemsHashs.
// * lh() means leafHash(x) := hash(0x00,x)
// * ih() means innerHash(x,y) := hash(0x01,x,y)
// * pb() means .PrimitiveBytes().
// * vp() means .ValuePreimage().Bytes().
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
	TypeID // never nil
	ValuePreimage
}

type ValuePreimage struct {
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
	buf := sizedBytes(tvp.TypeID[:])
	buf = append(buf, tvp.ValuePreimage.Bytes()...)
	return buf
}

func (vp ValuePreimage) String() string {
	return fmt.Sprintf("ValuePreimage{%X:%X:%d,%d,%d}",
		vp.ValType,
		vp.Data,
		vp.Offset,
		vp.Length,
		vp.Maxcap,
	)
}

func (vp *ValuePreimage) Bytes() []byte {
	buf := []byte{byte(vp.ValType)}
	switch vp.ValType {
	case ValTypeNil:
		return buf
	case ValTypePrimitive:
		fallthrough
	case ValTypePointer:
		fallthrough
	case ValTypeData:
		fallthrough
	case ValTypeObject:
		buf = append(buf, sizedBytes(vp.Data)...)
		return buf
	case ValTypeSlice:
		buf = append(buf, sizedBytes(vp.Data)...)
		buf = append(buf, uvarintBytes(uint64(vp.Offset))...)
		buf = append(buf, uvarintBytes(uint64(vp.Length))...)
		buf = append(buf, uvarintBytes(uint64(vp.Maxcap))...)
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
	if tep.TypeID.IsZero() {
		return nil
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
			if i == len(hz)-1 {
				h2 := Hashlet{} // zero
				hz[i/2] = innerHash(h1, h2)
			} else {
				h2 := hz[i+1]
				hz[i/2] = innerHash(h1, h2)
			}
		}
		// resize hz to half (or more).
		hz = hz[:(len(hz)+1)/2]
	}
	return hz[0]
}

//----------------------------------------
// Value.ValuePreimage
// Value.TypedElemPreimages

func (av *ArrayValue) ValuePreimage(
	rlm *Realm, owned bool) ValuePreimage {

	if av.Data != nil {
		// `ValuePreimage := 0x03,sz(data)` if byte-array.
		return ValuePreimage{
			ValType: ValTypeData,
			Data:    av.Data,
		}
	}
	// `ValuePreimage := 0x04,sz(ElemsHash)` if non-nil object.
	tepz := av.TypedElemPreimages(rlm, owned)
	eh := ElemsHashFromElements(tepz)
	return ValuePreimage{
		ValType:   ValTypeObject,
		Data:      eh[:],
		preimages: tepz,
	}
}

func (av *ArrayValue) TypedElemPreimages(
	rlm *Realm, owned bool) []TypedElemPreimage {

	avl := av.GetLength()
	if avl == 0 {
		return nil
	}
	// Sanity check.
	if av.Data != nil {
		panic("ArrayValue of data-bytes has no TypedElemPreimages," +
			"call ValuePreimage() instead")
	}
	// General (list) case.
	tepz := make([]TypedElemPreimage, avl)
	for i := 0; i < avl; i++ {
		ev := &av.List[i]
		tepz[i] = ev.TypedElemPreimage(rlm, owned)
	}
	return tepz
}

func (sv *SliceValue) ValuePreimage(
	rlm *Realm, owned bool) ValuePreimage {

	if sv.Base == nil {
		return ValuePreimage{} // nil slice
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
	// `ValuePreimage := 0x05,sz(vp(base)),off,len,max if slice.
	bvp := sv.Base.ValuePreimage(rlm, baseAsOwned)
	return ValuePreimage{
		ValType: ValTypeSlice,
		Data:    bvp.Bytes(),
		Offset:  sv.Offset,
		Length:  sv.Length,
		Maxcap:  sv.Maxcap,
	}
}

func (sv *StructValue) ValuePreimage(
	rlm *Realm, owned bool) ValuePreimage {

	// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
	tepz := sv.TypedElemPreimages(rlm, owned)
	eh := ElemsHashFromElements(tepz)
	return ValuePreimage{
		ValType:   ValTypeObject,
		Data:      eh[:],
		preimages: tepz,
	}
}

func (sv *StructValue) TypedElemPreimages(
	rlm *Realm, owned bool) []TypedElemPreimage {

	nf := len(sv.Fields)
	if nf == 0 {
		return nil
	}
	tvpz := make([]TypedElemPreimage, nf)
	for i := 0; i < nf; i++ {
		fv := &sv.Fields[i] // ref
		tvpz[i] = fv.TypedElemPreimage(rlm, owned)
	}
	return tvpz
}

func (mv *MapValue) ValuePreimage(
	rlm *Realm, owned bool) ValuePreimage {

	// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
	tepz := mv.TypedElemPreimages(rlm, owned)
	eh := ElemsHashFromElements(tepz)
	return ValuePreimage{
		ValType:   ValTypeObject,
		Data:      eh[:],
		preimages: tepz,
	}
}

// TODO split into keyOwned and valueOwned with new tags.
func (mv *MapValue) TypedElemPreimages(
	rlm *Realm, owned bool) []TypedElemPreimage {

	if mv.vmap == nil {
		return nil
	}

	ms := mv.List.Size
	tvpz := make([]TypedElemPreimage, ms*2) // even:key odd:val
	head := mv.List.Head
	// create deterministic list from mv.List
	for i := 0; i < ms; i++ {
		key, val := &head.Key, &head.Value // ref
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

// XXX dry code or something.
func (b *Block) ValuePreimage(
	rlm *Realm, owned bool) ValuePreimage {

	// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
	tepz := b.TypedElemPreimages(rlm, owned)
	eh := ElemsHashFromElements(tepz)
	return ValuePreimage{
		ValType:   ValTypeObject,
		Data:      eh[:],
		preimages: tepz,
	}
}

// XXX dry code, probably a method on TypedValuesList.
// XXX TODO: encode function values.
func (b *Block) TypedElemPreimages(
	rlm *Realm, owned bool) []TypedElemPreimage {

	nv := len(b.Values)
	if nv == 0 {
		return nil
	}
	tvpz := make([]TypedElemPreimage, nv)
	for i := 0; i < nv; i++ {
		tv := &b.Values[i] // ref
		switch baseOf(tv.T).(type) {
		case nil:
			// do nothing
		case *FuncType:
			// do nothing
		default:
			tvpz[i] = tv.TypedElemPreimage(rlm, owned)
		}
	}
	return tvpz
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
				TypeID: tid,
				ValuePreimage: ValuePreimage{
					ValType: ValTypePrimitive,
					Data:    pbz,
				},
			}
		} else {
			// `ValuePreimage := 0x00` if typed-nil.
			return TypedValuePreimage{
				TypeID: tid,
				ValuePreimage: ValuePreimage{
					ValType: ValTypeNil,
				},
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
					TypeID: tid,
					ValuePreimage: ValuePreimage{
						ValType: ValTypePointer,
						Data:    nil,
					},
				}
			} else {
				// `ValuePreimage := 0x02,sz(vh(*ptr))` if ptr.
				vh := tv.ValueHash(rlm, owned)
				return TypedValuePreimage{
					TypeID: tid,
					ValuePreimage: ValuePreimage{
						ValType: ValTypePointer,
						Data:    vh[:],
					},
				}
			}
		case *ArrayType:
			av := tv.V.(*ArrayValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			return TypedValuePreimage{
				TypeID:        tid,
				ValuePreimage: av.ValuePreimage(rlm, owned),
			}
		case *SliceType:
			sv := tv.V.(*SliceValue)
			// `ValuePreimage :=
			//    0x05,sz(vh(base)),off,len,max if slice.
			//  * TypeID is a byte-slice if byte-array.  NOTE:
			//  Slices do not have access to an underlying array's
			//  *DeclaredType if any.
			return TypedValuePreimage{
				TypeID:        tid,
				ValuePreimage: sv.ValuePreimage(rlm, owned),
			}
		case *StructType:
			sv := tv.V.(*StructValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			return TypedValuePreimage{
				TypeID:        tid,
				ValuePreimage: sv.ValuePreimage(rlm, owned),
			}
		case *MapType:
			mv := tv.V.(*StructValue)
			// `ValuePreimage := 0x04,sz(ElemsHash)` if object.
			return TypedValuePreimage{
				TypeID:        tid,
				ValuePreimage: mv.ValuePreimage(rlm, owned),
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
