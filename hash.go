package gno

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
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
	res = HashBytes(buf)
	return
}

func innerHash(h1, h2 Hashlet) (res Hashlet) {
	buf := make([]byte, 1+HashSize*2)
	buf[0] = 0x01
	copy(buf[1:1+HashSize], h1[:])
	copy(buf[1+HashSize:], h2[:])
	res = HashBytes(buf)
	return
}

//----------------------------------------
// ValueHash
//
// The ValueHash of a typed value is a unique deterministic
// accountable fingerprint of that typed value, and can be used
// to prove the value or any part of its value which is
// accessible from the Gno language.
//
// `ValueHash := lh(ValueImage)`
// `ValueImage := 0x00` if nil value.
// `ValueImage := 0x01,varint(.) if fixed-numeric.
// `ValueImage := 0x02,sz(bytes)` if variable length bytes.
// `ValueImage := 0x03,sz(TypeID),vi(*ptr)` if non-nil ptr.
// `ValueImage := 0x04,sz(OwnerID),sz(ElemsHash),ref` if object.
// `ValueImage := 0x05,vi(base),off,len,max if slice.
// `ValueImage := 0x06,sz(TypeID)` if type.
//
// `ElemsHash := lh(ElemImage)` if object w/ 1 elem.
// `ElemsHash := ih(eh(Left),eh(Right))` if object w/ 2+ elems.
//
// `ElemImage:`
//   `= 0x10` if nil interface.
//   `= 0x11,sz(ObjectID),sz(TypeID)` if borrowed.
//   `= 0x12,sz(ObjectID),sz(TypedValueHash)` if owned.
//   `= 0x13,sz(TypeID),sz(ValueHash)` if other.
//    - other: prim/ptr/slice/type/typed-nil.
//    - ownership passed through for pointers/slices/arrays.
//
// `TypedValueHash := lh(sz(TypeID),sz(ValueHash))`
//
// * eh() are inner ElemsHashs.
// * lh() means leafHash(x) := hash(0x00,x)
// * ih() means innerHash(x,y) := hash(0x01,x,y)
// * pb() means .PrimitiveBytes().
// * sz() means (varint) size-prefixed bytes.
// * vi() means .ValueImage().Bytes().
// * off,len,max and other integers are varint encoded.
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
// any of the elements have refcount>1, image derivation
// panics.

type ValueHash Hashlet

func (vh ValueHash) IsZero() bool {
	return vh == (ValueHash{})
}

func (vh ValueHash) Bytes() []byte {
	return vh[:]
}

type TypedValueHash Hashlet

func (tvh TypedValueHash) IsZero() bool {
	return tvh == (TypedValueHash{})
}

func (tvh TypedValueHash) Bytes() []byte {
	return tvh[:]
}

func DeriveTypedValueHash(typeID TypeID, vh ValueHash) TypedValueHash {
	bz := sizedBytes(typeID.Bytes())
	bz = append(bz, sizedBytes(vh.Bytes())...)
	return TypedValueHash(leafHash(bz))
}

//----------------------------------------
// ValueImage

type ValType byte

const (
	ValTypeNil     = ValType(0x00)
	ValTypeNumeric = ValType(0x01)
	ValTypeBytes   = ValType(0x02)
	ValTypePointer = ValType(0x03)
	ValTypeObject  = ValType(0x04)
	ValTypeSlice   = ValType(0x05)
	ValTypeType    = ValType(0x06)
)

type ValueImage struct {
	ValType                // 1:num,2:bz,3:ptr,4:obj,5:sli,6:ty
	Data       []byte      // if primitive, data-array, obj
	TypeID                 // if non-nil ptr or type
	Base       *ValueImage // if non-nil ptr or slice
	Offset     int         // if slice
	Length     int         // if slice
	Maxcap     int         // if slice
	OwnerID    ObjectID    // if obj
	ElemImages []ElemImage // if obj; for persistence
	RefCount   int         // if obj
}

func (vi *ValueImage) IsZero() bool {
	return vi == nil
}

func (vi *ValueImage) String() string {
	return vi.StringWithElems(true)
}

func (vi *ValueImage) StringWithElems(withElems bool) string {
	if vi == nil {
		return "VI[nil]"
	}
	switch vi.ValType {
	case ValTypeNumeric:
		return fmt.Sprintf("VI[numeric:%X]",
			vi.Data,
		)
	case ValTypeBytes:
		return fmt.Sprintf("VI[data:0x%X]",
			vi.Data,
		)
	case ValTypePointer:
		return fmt.Sprintf("VI[pointer:%s,%s]",
			vi.TypeID.String(),
			vi.Base.String(),
		)
	case ValTypeObject:
		if withElems {
			pz := []string{}
			for _, image := range vi.ElemImages {
				pz = append(pz, "- "+image.String())
			}
			return fmt.Sprintf("VI[object:%s#%X&%d]:\n%s",
				vi.OwnerID.String(),
				vi.Data,
				vi.RefCount,
				strings.Join(pz, "\n"),
			)
		} else {
			return fmt.Sprintf("VI[object:%s#%X&%d]",
				vi.OwnerID.String(),
				vi.Data,
				vi.RefCount)
		}
	case ValTypeSlice:
		return fmt.Sprintf("VI[slice:%s[%d,%d,%d]]",
			vi.Base.String(),
			vi.Offset,
			vi.Length,
			vi.Maxcap,
		)
	case ValTypeType:
		return fmt.Sprintf("VI[type:%s]",
			vi.TypeID.String(),
		)
	default:
		panic("should not happen")
	}
}

func (vi *ValueImage) Bytes() []byte {
	if vi == nil {
		return []byte{byte(ValTypeNil)} // Special case.
	}
	buf := []byte{byte(vi.ValType)}
	switch vi.ValType {
	case ValTypeNumeric:
		buf = append(buf, vi.Data...)
		return buf
	case ValTypeBytes:
		buf = append(buf, sizedBytes(vi.Data)...)
		return buf
	case ValTypePointer:
		buf = append(buf, sizedBytes(vi.TypeID.Bytes())...)
		buf = append(buf, vi.Base.Bytes()...)
		return buf
	case ValTypeObject:
		buf = append(buf, sizedBytes(vi.OwnerID.Bytes())...)
		buf = append(buf, sizedBytes(vi.Data)...)
		buf = append(buf, varintBytes(int64(vi.RefCount))...)
		return buf
	case ValTypeSlice:
		buf = append(buf, vi.Base.Bytes()...)
		buf = append(buf, varintBytes(int64(vi.Offset))...)
		buf = append(buf, varintBytes(int64(vi.Length))...)
		buf = append(buf, varintBytes(int64(vi.Maxcap))...)
		return buf
	case ValTypeType:
		buf = append(buf, sizedBytes(vi.TypeID.Bytes())...)
		return buf
	default:
		panic("should not happen")
	}
}

func (vi *ValueImage) ValueHash() ValueHash {
	return ValueHash(leafHash(vi.Bytes()))
}

//----------------------------------------
// ElemImage

type ElemImage struct {
	ElemType       // 0x10:nil,0x11:brwd,0x12:owned,0x13:other
	ObjectID       // if borrowed or owned
	TypedValueHash // if owned
	TypeID         // if borrowed or other
	ValueHash      // if other (prim/ptr/slice/typed-nil)
	*ValueImage    // if owned or other
}

type ElemType byte

const (
	ElemTypeInvalid  = ElemType(0x00)
	ElemTypeNil      = ElemType(0x10)
	ElemTypeBorrowed = ElemType(0x11)
	ElemTypeOwned    = ElemType(0x12)
	ElemTypeOther    = ElemType(0x13)
)

func (ei *ElemImage) String() string {
	switch ei.ElemType {
	case ElemTypeNil:
		return "EI[nil]"
	case ElemTypeBorrowed:
		return fmt.Sprintf(
			"EI[%s:%s(.)]",
			ei.ObjectID.String(),
			ei.TypeID.String())
	case ElemTypeOwned:
		return fmt.Sprintf(
			"EI[%s:#%X] // TypeID:%s %s",
			ei.ObjectID.String(),
			ei.TypedValueHash.Bytes(),
			ei.TypeID.String(),
			ei.ValueImage.StringWithElems(false))
	case ElemTypeOther:
		return fmt.Sprintf(
			"EI[:%s(#%X)] // %s",
			ei.TypeID.String(),
			ei.ValueHash.Bytes(),
			ei.ValueImage.StringWithElems(false))
	default:
		panic("should not happen")
	}
}

func (ei *ElemImage) Bytes() []byte {
	buf := []byte{byte(ei.ElemType)}
	switch ei.ElemType {
	case ElemTypeNil:
		if debug {
			if !ei.ObjectID.IsZero() {
				panic("should not happen")
			}
			if !ei.TypedValueHash.IsZero() {
				panic("should not happen")
			}
			if !ei.TypeID.IsZero() {
				panic("should not happen")
			}
			if !ei.ValueHash.IsZero() {
				panic("should not happen")
			}
			if !ei.ValueImage.IsZero() {
				panic("should not happen")
			}
		}
		return buf
	case ElemTypeBorrowed:
		if debug {
			if ei.ObjectID.IsZero() {
				panic("should not happen")
			}
			if !ei.TypedValueHash.IsZero() {
				panic("should not happen")
			}
			if ei.TypeID.IsZero() {
				panic("should not happen")
			}
			if !ei.ValueHash.IsZero() {
				panic("should not happen")
			}
			if !ei.ValueImage.IsZero() {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(ei.ObjectID.Bytes())...)
		buf = append(buf, sizedBytes(ei.TypeID.Bytes())...)
		return buf
	case ElemTypeOwned:
		if debug {
			if ei.ObjectID.IsZero() {
				panic("should not happen")
			}
			if ei.TypedValueHash.IsZero() {
				panic("should not happen")
			}
			if ei.TypeID.IsZero() {
				panic("should not happen")
			}
			if ei.ValueHash.IsZero() {
				panic("should not happen")
			}
			if ei.ValueImage.IsZero() {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(ei.ObjectID.Bytes())...)
		buf = append(buf, sizedBytes(ei.TypeID.Bytes())...)
		buf = append(buf, sizedBytes(ei.ValueHash.Bytes())...)
		return buf
	case ElemTypeOther:
		if debug {
			if !ei.ObjectID.IsZero() {
				panic("should not happen")
			}
			if !ei.TypedValueHash.IsZero() {
				panic("should not happen")
			}
			if ei.TypeID.IsZero() {
				panic("should not happen")
			}
			if ei.ValueHash.IsZero() {
				panic("should not happen")
			}
			if ei.ValueImage.IsZero() {
				panic("should not happen")
			}
		}
		buf = append(buf, sizedBytes(ei.TypeID.Bytes())...)
		buf = append(buf, ei.ValueImage.Bytes()...)
		return buf
	default:
		panic("should not happen")
	}
}

func (ei *ElemImage) LeafHash() Hashlet {
	return leafHash(ei.Bytes())
}

func ElemsHashFromElements(eiz []ElemImage) Hashlet {
	// special case if nil
	if eiz == nil {
		return Hashlet{}
	}
	// translate to leaf hashes
	hz := make([]Hashlet, ((len(eiz)+1)/2)*2)
	for i, tvi := range eiz {
		hz[i] = tvi.LeafHash()
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
// Object.ValueImage
// Object.ElemImages

func (av *ArrayValue) ValueImage(
	rlm *Realm, owned bool) (vi *ValueImage) {

	// Create or update when deriving
	// the value images of an owned object.
	if owned {
		defer func() {
			rlm.maybeSaveObject(av, vi)
		}()
	}
	// `ValueImage := 0x02,sz(bytes)` if variable length bytes.
	if av.Data != nil {
		vi = &ValueImage{
			ValType: ValTypeBytes,
			Data:    av.Data,
		}
		return
	}
	// `ValueImage :=
	//   0x04,sz(OwnerID),sz(ElemsHash),ref` if object.
	eiz := av.ElemImages(rlm, owned)
	eh := ElemsHashFromElements(eiz)
	return &ValueImage{
		ValType:    ValTypeObject,
		OwnerID:    av.GetOwnerID(),
		Data:       eh[:],
		RefCount:   av.GetRefCount(),
		ElemImages: eiz,
	}
}

func (av *ArrayValue) ElemImages(
	rlm *Realm, owned bool) []ElemImage {

	avl := av.GetLength()
	if avl == 0 {
		return nil
	}
	// Sanity check.
	if av.Data != nil {
		panic("ArrayValue of data-bytes has no ElemImages," +
			"call ValueImage() instead")
	}
	// General (list) case.
	eiz := make([]ElemImage, avl)
	for i := 0; i < avl; i++ {
		ev := &av.List[i]
		eiz[i] = ev.ElemImage(rlm, owned)
	}
	return eiz
}

func (sv *SliceValue) ValueImage(
	rlm *Realm, owned bool) *ValueImage {

	if sv.Base == nil {
		return nil
	}
	// If (self, base) is:
	//  - (owned, already-owned):
	//    * panic (ownership conflict)
	//  - (owned, not-owned):
	//    * TODO: trim array to slice window first
	//    * save & hash array w/ its object_id
	//    * leafHash(Base.TVI(true),Offset,Length,Maxcap)
	//  - (not-owned, already-owned):
	//    * leafHash(Base.TVI(false),Offset,Length,Maxcap)
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
	// `ValueImage := 0x05,vi(base),off,len,max if slice.
	bvi := sv.Base.ValueImage(rlm, baseAsOwned)
	return &ValueImage{
		ValType: ValTypeSlice,
		Base:    bvi,
		Offset:  sv.Offset,
		Length:  sv.Length,
		Maxcap:  sv.Maxcap,
	}
}

func (sv *StructValue) ValueImage(
	rlm *Realm, owned bool) (vi *ValueImage) {

	// Create or update when deriving
	// the value images of an owned object.
	if owned {
		defer func() {
			rlm.maybeSaveObject(sv, vi)
		}()
	}
	// `ValueImage :=
	//   0x04,sz(OwnerID),sz(ElemsHash),ref` if object.
	eiz := sv.ElemImages(rlm, owned)
	eh := ElemsHashFromElements(eiz)
	return &ValueImage{
		ValType:    ValTypeObject,
		RefCount:   sv.GetRefCount(),
		Data:       eh[:],
		OwnerID:    sv.GetOwnerID(),
		ElemImages: eiz,
	}
	return
}

func (sv *StructValue) ElemImages(
	rlm *Realm, owned bool) []ElemImage {

	nf := len(sv.Fields)
	if nf == 0 {
		return nil
	}
	tviz := make([]ElemImage, nf)
	for i := 0; i < nf; i++ {
		fv := &sv.Fields[i] // ref
		tviz[i] = fv.ElemImage(rlm, owned)
	}
	return tviz
}

func (mv *MapValue) ValueImage(
	rlm *Realm, owned bool) (vi *ValueImage) {

	// Create or update when deriving
	// the value images of an owned object.
	if owned {
		defer func() {
			rlm.maybeSaveObject(mv, vi)
		}()
	}
	// `ValueImage :=
	//   0x04,sz(OwnerID),sz(ElemsHash),ref` if object.
	eiz := mv.ElemImages(rlm, owned)
	eh := ElemsHashFromElements(eiz)
	return &ValueImage{
		ValType:    ValTypeObject,
		OwnerID:    mv.GetOwnerID(),
		Data:       eh[:],
		RefCount:   mv.GetRefCount(),
		ElemImages: eiz,
	}
}

// TODO split into keyOwned and valueOwned with new tags.
func (mv *MapValue) ElemImages(
	rlm *Realm, owned bool) []ElemImage {

	if mv.vmap == nil {
		return nil
	}

	ms := mv.List.Size
	tviz := make([]ElemImage, ms*2) // even:key odd:val
	head := mv.List.Head
	// create deterministic list from mv.List
	for i := 0; i < ms; i++ {
		key, val := &head.Key, &head.Value // ref
		tviz[i*2+0] = key.ElemImage(rlm, owned)
		tviz[i*2+1] = val.ElemImage(rlm, owned)
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
	return tviz
}

func (tv *TypeValue) ValueImage(
	rlm *Realm, owned bool) *ValueImage {

	// `ValueImage := 0x06,sz(TypeID)` if type.
	return &ValueImage{
		ValType: ValTypeType,
		TypeID:  tv.Type.TypeID(),
	}
}

// XXX dry code or something.
func (b *Block) ValueImage(
	rlm *Realm, owned bool) (vi *ValueImage) {

	// Create or update when deriving
	// the value images of an owned object.
	if owned {
		defer func() {
			rlm.maybeSaveObject(b, vi)
		}()
	}
	// `ValueImage :=
	//   0x04,sz(ObjectInfo),sz(ElemsHash)` if object.
	eiz := b.ElemImages(rlm, owned)
	eh := ElemsHashFromElements(eiz)
	return &ValueImage{
		ValType:    ValTypeObject,
		OwnerID:    b.GetOwnerID(),
		Data:       eh[:],
		RefCount:   b.GetRefCount(),
		ElemImages: eiz,
	}
}

// XXX dry code, probably a method on TypedValuesList.
// XXX TODO: encode function values.
func (b *Block) ElemImages(
	rlm *Realm, owned bool) []ElemImage {

	nv := len(b.Values)
	if nv == 0 {
		return nil
	}
	tviz := make([]ElemImage, nv)
	for i := 0; i < nv; i++ {
		tv := &b.Values[i] // ref
		switch baseOf(tv.T).(type) {
		case *FuncType:
			// XXX replace this with real image.
			tviz[i] = ElemImage{
				ElemType: ElemTypeNil,
			}
		default:
			tviz[i] = tv.ElemImage(rlm, owned)
		}
	}
	return tviz
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
		vi := tv.ValueImage(rlm, owned)
		// `ValueHash := lh(ValueImage)`
		return vi.ValueHash()
	}
}

// Any dirty or new-real value will be saved to realm.
func (tv *TypedValue) ValueImage(rlm *Realm, owned bool) *ValueImage {
	if tv.IsUndefined() {
		// `ValueImage := 0x00` if nil value.
		return nil
	}
	if tv.V == nil { // primitive or nil
		if _, ok := baseOf(tv.T).(PrimitiveType); ok {
			pbz, isVarint := tv.PrimitiveBytes()
			if !isVarint {
				panic("should not happen")
			}
			// `ValueImage := 0x01,varint(.) if fixed-numeric.
			return &ValueImage{
				ValType: ValTypeNumeric,
				Data:    pbz,
			}
		} else {
			// `ValueImage := 0x00` if nil value.
			return nil // 0x00 signified with nil *ValueIamge.
		}
	} else { // non-nil object.
		switch baseOf(tv.T).(type) {
		case PrimitiveType:
			pbz, isVarint := tv.PrimitiveBytes()
			if isVarint {
				panic("should not happen")
			}
			// `ValueImage := 0x02,sz(bytes)` if size-prefixed bytes.
			return &ValueImage{
				ValType: ValTypeBytes,
				Data:    pbz,
			}
		case PointerType:
			pv := tv.V.(PointerValue)
			if pv.TypedValue == nil {
				panic("should not happen")
			} else {
				// vi := 0x03,sz(TypeID),vi(*ptr) if non-nil ptr.
				pvi := pv.ValueImage(rlm, owned)
				ptid := tv.T.TypeID()
				return &ValueImage{
					ValType: ValTypePointer,
					TypeID:  ptid,
					Base:    pvi,
				}
			}
		case *ArrayType:
			av := tv.V.(*ArrayValue)
			return av.ValueImage(rlm, owned)
		case *SliceType:
			sv := tv.V.(*SliceValue)
			return sv.ValueImage(rlm, owned)
		case *StructType:
			sv := tv.V.(*StructValue)
			return sv.ValueImage(rlm, owned)
		case *MapType:
			mv := tv.V.(*StructValue)
			return mv.ValueImage(rlm, owned)
		case *TypeType:
			t := tv.GetType()
			// `ValueImage := 0x06,sz(TypeID)` if type.
			return &ValueImage{
				ValType: ValTypeType,
				TypeID:  t.TypeID(),
			}
		default:
			panic(fmt.Sprintf(
				"unexpected type for ValueImage(): %s",
				tv.T.String()))
		}
	}
}

// Main entrypoint for objects to get the EI of elements.
// Any dirty or new-real elements will be saved to realm.
func (tv *TypedValue) ElemImage(rlm *Realm, owned bool) ElemImage {
	if tv.IsUndefined() {
		// `ElemImage := 0x10` if nil interface.
		return ElemImage{
			ElemType: ElemTypeNil,
		}
	} else if tv.T.Kind() == InterfaceKind {
		if debug {
			panic("should not happen")
		}
	}
	if tv.V == nil {
		// 0x13,sz(TypeID),sz(ValueHash) if other.
		vi := tv.ValueImage(rlm, owned)
		return ElemImage{
			ElemType:   ElemTypeOther,
			TypeID:     tv.T.TypeID(),
			ValueHash:  vi.ValueHash(),
			ValueImage: vi,
		}
	} else {
		switch baseOf(tv.T).(type) {
		case PrimitiveType:
			// 0x13,sz(TypeID),sz(ValueHash) if other.
			vi := tv.ValueImage(rlm, owned)
			return ElemImage{
				ElemType:   ElemTypeOther,
				TypeID:     tv.T.TypeID(),
				ValueHash:  vi.ValueHash(),
				ValueImage: vi,
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
			if owned {
				// 0x12,sz(ObjectID),sz(TypedValueHash)` if owned
				// TypedValueHash := lh(sz(TypeID),sz(ValueHash))
				tid := tv.T.TypeID()
				vi := tv.ValueImage(rlm, owned)
				vh := vi.ValueHash()
				tvh := DeriveTypedValueHash(tid, vh)
				return ElemImage{
					ElemType:       ElemTypeOwned,
					ObjectID:       obj.MustGetObjectID(),
					TypeID:         tid,
					TypedValueHash: tvh,
					ValueHash:      vh,
					ValueImage:     vi,
				}
			} else {
				// 0x11,sz(ObjectID),sz(TypeID) if borrowed.
				return ElemImage{
					ElemType: ElemTypeBorrowed,
					ObjectID: obj.MustGetObjectID(),
					TypeID:   tv.T.TypeID(),
				}
			}
		case PointerType, *SliceType, *TypeType:
			// 0x13,sz(TypeID),sz(ValueHash) if other.
			vi := tv.ValueImage(rlm, owned)
			return ElemImage{
				ElemType:   ElemTypeOther,
				TypeID:     tv.T.TypeID(),
				ValueHash:  vi.ValueHash(),
				ValueImage: vi,
			}
		default:
			panic(fmt.Sprintf(
				"unexpected type for elem images: %s",
				tv.T.String()))
		}
	}
}

//----------------------------------------
// misc

func varintBytes(u int64) []byte {
	var buf [10]byte
	n := binary.PutVarint(buf[:], u)
	return buf[0:n]
}

func sizedBytes(bz []byte) []byte {
	bz2 := make([]byte, len(bz)+10)
	n := binary.PutVarint(bz2[:10], int64(len(bz)))
	copy(bz2[n:n+len(bz)], bz)
	return bz2[:n+len(bz)]
}
