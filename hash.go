package gno

import (
	"crypto/sha256"
	"encoding/binary"
)

const HashSize = 20

type Hashlet [HashSize]byte

func (h Hashlet) Bytes() []byte {
	return h[:]
}

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
// TypedValueImage, etc.

type TypedValueImage struct {
	TypeID
	ValueImage
}

type ValueHash struct {
	Hashlet
}

type ValueImage interface {
	assertValueImage()
	//String() string
}

func (_ ObjectInfoImage) assertValueImage()       {}
func (_ WeakRefValueImage) assertValueImage()     {}
func (_ PrimitiveValueImage) assertValueImage()   {}
func (_ PointerValueImage) assertValueImage()     {}
func (_ ArrayValueImage) assertValueImage()       {}
func (_ SliceValueImage) assertValueImage()       {}
func (_ StructValueImage) assertValueImage()      {}
func (_ FuncValueImage) assertValueImage()        {}
func (_ BoundMethodValueImage) assertValueImage() {}
func (_ MapValueImage) assertValueImage()         {}
func (_ TypeValueImage) assertValueImage()        {}
func (_ PackageValueImage) assertValueImage()     {}
func (_ BlockValueImage) assertValueImage()       {}

type ObjectInfoImage struct {
	ID       ObjectID
	OwnerID  ObjectID
	ModTime  uint64
	RefCount int
}

type WeakRefValueImage struct {
	ID ObjectID
}

type PrimitiveValueImage []byte

type PointerValueImage struct {
	// BaseID           ObjectID // if weak
	// Index            int      // if weak
	*TypedValueImage // if owned
}

type ArrayValueImage struct {
	ObjectInfo ObjectInfoImage
	List       []TypedValueImage
	Data       []byte
}

type SliceValueImage struct {
	// BaseID           ObjectID // if weak
	*ArrayValueImage // if owned
	Offset           int
	Length           int
	Maxcap           int
}

type StructValueImage struct {
	ObjectInfo ObjectInfoImage
	Fields     []TypedValueImage
}

type FuncValueImage struct {
	Type     TypeID
	IsMethod bool
	Name     Name
	Closure  BlockValueImage
	FileName Name
	PkgPath  string
}

type BoundMethodValueImage struct {
	Func     FuncValueImage
	Receiver TypedValueImage
}

type MapValueImage struct {
	ObjectInfo ObjectInfoImage
	List       []MapItemImage
}

type MapItemImage struct {
	Key   TypedValueImage
	Value TypedValueImage
}

type TypeValueImage struct {
	TypeID
}

type PackageValueImage struct {
	Block   BlockValueImage
	PkgName Name
	PkgPath string
}

type BlockValueImage struct {
	ObjectInfo ObjectInfoImage
	ParentID   ObjectID
	Values     []TypedValueImage
}

//----------------------------------------

type ImageEncoder struct {
	TypeLookup    func(TypeID) Type
	PackageLookup func(pkgPath string) *PackageValue
}

func (ie ImageEncoder) EncodeTypedValueImage(tv TypedValue) TypedValueImage {
	typeid := tv.T.TypeID()
	valimg := ie.EncodeValueImage(tv)
	return TypedValueImage{
		TypeID:     typeid,
		ValueImage: valimg,
	}
}

func (ie ImageEncoder) EncodeValueImage(tv TypedValue) ValueImage {
	switch baseOf(tv.T).(type) {
	case PrimitiveType:
		return PrimitiveValueImage(tv.PrimitiveBytes())
	case *PointerType:
		if tv.V == nil {
			return PointerValueImage{}
		} else {
			val := tv.V.(PointerValue)
			tvi := ie.EncodeTypedValueImage(*val.TypedValue)
			return PointerValueImage{
				TypedValueImage: &tvi,
			}
		}
	case *ArrayType:
		if tv.V == nil {
			return ArrayValueImage{}
		} else {
			val := tv.V.(*ArrayValue)
			valimg := ArrayValueImage{}
			valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
			if val.Data == nil {
				valimg.List = make([]TypedValueImage, len(val.List))
				for i, item := range val.List {
					valimg.List[i] = ie.EncodeTypedValueImage(item)
				}
			} else {
				valimg.Data = make([]byte, len(val.Data))
				copy(valimg.Data, val.Data)
			}
			return valimg
		}
	case *SliceType:
		if tv.V == nil {
			return SliceValueImage{}
		} else {
			val := tv.V.(*SliceValue)
			valimg := SliceValueImage{}
			avi := ie.EncodeValueImage(TypedValue{
				T: &ArrayType{}, // XXX hack
				V: val.Base,
			}).(ArrayValueImage)
			valimg.ArrayValueImage = &avi
			valimg.Offset = val.Offset
			valimg.Length = val.Length
			valimg.Maxcap = val.Maxcap
			return valimg
		}
	case *StructType:
		if tv.V == nil {
			return StructValueImage{}
		} else {
			val := tv.V.(*StructValue)
			valimg := StructValueImage{}
			valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
			valimg.Fields = make([]TypedValueImage, len(val.Fields))
			for i, field := range val.Fields {
				valimg.Fields[i] = ie.EncodeTypedValueImage(field)
			}
			return valimg
		}
	case *FuncType:
		panic("not yet supported")
	case *MapType:
		if tv.V == nil {
			return MapValueImage{}
		} else {
			val := tv.V.(*MapValue)
			valimg := MapValueImage{}
			valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
			valimg.List = make([]MapItemImage, 0, val.List.Size)
			for cur := val.List.Head; cur != nil; cur = cur.Next {
				valimg.List = append(valimg.List, MapItemImage{
					Key:   ie.EncodeTypedValueImage(cur.Key),
					Value: ie.EncodeTypedValueImage(cur.Value),
				})
			}
			return valimg
		}
	case *InterfaceType:
		panic("should not happen")
	case *TypeType:
		panic("not yet supported")
	case *PackageType:
		val := tv.V.(*PackageValue)
		valimg := PackageValueImage{}
		valimg.Block = ie.EncodeValueImage(TypedValue{
			T: &blockType{},
			V: blockValue{&val.Block},
		}).(BlockValueImage)
		valimg.PkgName = val.PkgName
		valimg.PkgPath = val.PkgPath
		return valimg
	case *ChanType:
		panic("not yet supported")
	case *nativeType:
		panic("not yet supported") // maybe never will.
	case blockType:
		val := tv.V.(blockValue)
		valimg := BlockValueImage{}
		valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
		valimg.ParentID = val.Parent.ID
		valimg.Values = make([]TypedValueImage, len(val.Values))
		for i, tv := range val.Values {
			valimg.Values[i] = ie.EncodeTypedValueImage(tv)
		}
		return valimg
	default:
		panic("should not happen")
	}
}

func encodeObjectInfo(oi ObjectInfo) ObjectInfoImage {
	return ObjectInfoImage{
		ID:       oi.ID,
		OwnerID:  oi.OwnerID,
		ModTime:  oi.ModTime,
		RefCount: oi.RefCount,
	}
}

func (ie ImageEncoder) DecodeTypedValueImage(tvi TypedValueImage) TypedValue {
	// XXX what else?
	return TypedValue{}
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

func isASCIIText(bz []byte) bool {
	if len(bz) == 0 {
		return false
	}
	for _, b := range bz {
		if 32 <= b && b <= 126 {
			// good
		} else {
			return false
		}
	}
	return true
}
