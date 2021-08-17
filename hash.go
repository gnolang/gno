package gno

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/gnolang/gno/pkgs/amino"
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
	TypeID     TypeID
	ValueImage ValueImage
}

type ValueHash struct {
	Hashlet
}

type ValueImage interface {
	assertValueImage()
	//String() string
}

func (_ ObjectInfoImage) assertValueImage()       {}
func (_ RefImage) assertValueImage()              {}
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
	_RealmID      RealmID
	NewTime       uint64 // of ID
	_OwnerNewTime uint64 // of ID
	_ModTime      uint64
	_RefCount     int
}

type RefImage struct {
	RealmID RealmID   // if cross-realm
	NewTime uint64    // required
	Hash    ValueHash // if owned
}

type PrimitiveValueImage []byte

type PointerValueImage struct {
	// BaseID           ObjectID // if weak
	// Index            int      // if weak
	TypedValue TypedValueImage // if owned
}

type ArrayValueImage struct {
	ObjectInfo ObjectInfoImage
	List       []TypedValueImage
	Data       []byte
}

type SliceValueImage struct {
	// BaseID           ObjectID // if weak
	ArrayRef RefImage // if owned
	Offset   int
	Length   int
	Maxcap   int
}

type StructValueImage struct {
	ObjectInfo ObjectInfoImage
	Fields     []TypedValueImage
}

type FuncValueImage struct {
	TypeID     TypeID
	IsMethod   bool
	Name       Name
	ClosureRef RefImage
	FileName   Name
	PkgPath    string
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
	TypeID TypeID
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

func hashValueImage(vi ValueImage) ValueHash {
	bz := amino.MustMarshal(vi)
	return ValueHash{HashBytes(bz)}
}

//----------------------------------------
// ImageCodec

type ImageCodec struct {
	RealmID       RealmID
	TypeLookup    func(TypeID) Type
	PackageLookup func(pkgPath string) *PackageValue
}

func (ic ImageCodec) EncodeTypedValueImage(tv TypedValue) TypedValueImage {
	if tv.IsUndefined() {
		return TypedValueImage{}
	} else {
		typeid := tv.T.TypeID()
		valimg := ic.EncodeValueImage(tv)
		return TypedValueImage{
			TypeID:     typeid,
			ValueImage: valimg,
		}
	}
}

func (ic ImageCodec) EncodeTypedRefImage(tv TypedValue) TypedValueImage {
	typeid := tv.T.TypeID()
	refimg := ic.EncodeRefImage(tv.V.(Object))
	return TypedValueImage{
		TypeID:     typeid,
		ValueImage: refimg,
	}
}

func (ic ImageCodec) EncodeRefImage(oo Object) RefImage {
	if debug {
		if ic.RealmID.IsZero() {
			panic("should not happen")
		}
	}
	if oo == nil {
		panic("should not happen")
	}
	oi := oo.GetObjectInfo()
	if debug {
		if oi.ID.IsZero() {
			panic("should not happen")
		}
		if oi.Hash.IsZero() {
			panic("should not happen")
		}
		if oi.RefCount != 1 {
			panic("should not happen")
		}
	}
	if oi.ID.RealmID == ic.RealmID {
		return RefImage{
			NewTime: oi.ID.NewTime,
			Hash:    oi.Hash,
		}
	} else {
		return RefImage{
			RealmID: oi.ID.RealmID,
			NewTime: oi.ID.NewTime,
			Hash:    oi.Hash,
		}
	}
}

func (ic ImageCodec) EncodeObjectImage(oo Object) ValueImage {
	switch val := oo.(type) {
	case nil:
		panic("should not happen")
	case *ArrayValue:
		valimg := ArrayValueImage{}
		valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
		if val.Data == nil {
			valimg.List = make([]TypedValueImage, len(val.List))
			for i, item := range val.List {
				valimg.List[i] = ic.EncodeTypedValueImage(item)
			}
		} else {
			valimg.Data = make([]byte, len(val.Data))
			copy(valimg.Data, val.Data)
		}
		return valimg
	case *StructValue:
		valimg := StructValueImage{}
		valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
		valimg.Fields = make([]TypedValueImage, len(val.Fields))
		for i, field := range val.Fields {
			valimg.Fields[i] = ic.EncodeTypedValueImage(field)
		}
		return valimg
	case *MapValue:
		valimg := MapValueImage{}
		valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
		valimg.List = make([]MapItemImage, 0, val.List.Size)
		for cur := val.List.Head; cur != nil; cur = cur.Next {
			valimg.List = append(valimg.List, MapItemImage{
				Key:   ic.EncodeTypedValueImage(cur.Key),
				Value: ic.EncodeTypedValueImage(cur.Value),
			})
		}
		return valimg
	case *Block:
		valimg := BlockValueImage{}
		valimg.ObjectInfo = encodeObjectInfo(val.ObjectInfo)
		if val.Parent != nil {
			valimg.ParentID = val.Parent.ID
		}
		valimg.Values = make([]TypedValueImage, len(val.Values))
		for i, tv := range val.Values {
			valimg.Values[i] = ic.EncodeTypedValueImage(tv)
		}
		return valimg
	default:
		panic(fmt.Sprintf(
			"unexpected value type %v",
			reflect.TypeOf(oo)))
	}
}

func (ic ImageCodec) EncodeValueImage(tv TypedValue) ValueImage {
	switch baseOf(tv.T).(type) {
	case PrimitiveType:
		return PrimitiveValueImage(tv.PrimitiveBytes())
	case *PointerType:
		if tv.V == nil {
			return PointerValueImage{}
		} else {
			val := tv.V.(PointerValue)
			tvi := ic.EncodeTypedRefImage(*val.TypedValue)
			return PointerValueImage{
				TypedValue: tvi,
			}
		}
	case *SliceType:
		if tv.V == nil {
			return SliceValueImage{}
		} else {
			val := tv.V.(*SliceValue)
			valimg := SliceValueImage{}
			valimg.ArrayRef = ic.EncodeRefImage(val.Base)
			valimg.Offset = val.Offset
			valimg.Length = val.Length
			valimg.Maxcap = val.Maxcap
			return valimg
		}
	case *FuncType:
		if tv.V == nil {
			return FuncValueImage{}
		} else {
			val := tv.V.(*FuncValue)
			valimg := FuncValueImage{}
			valimg.TypeID = val.Type.TypeID()
			valimg.IsMethod = val.IsMethod
			valimg.Name = val.Name
			if val.Closure != nil {
				valimg.ClosureRef = ic.EncodeRefImage(val.Closure)
			}
			valimg.FileName = val.FileName
			valimg.PkgPath = val.pkg.PkgPath
			return valimg
		}
	case *InterfaceType:
		panic("should not happen")
	case *TypeType:
		if tv.V == nil {
			return TypeValueImage{}
		} else {
			return TypeValueImage{TypeID: tv.GetType().TypeID()}
		}
	case *PackageType:
		val := tv.V.(*PackageValue)
		valimg := PackageValueImage{}
		valimg.Block = ic.EncodeValueImage(TypedValue{
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
	case *ArrayType:
		if tv.V == nil {
			return ArrayValueImage{}
		} else {
			return ic.EncodeObjectImage(tv.V.(Object))
		}
	case *StructType:
		if tv.V == nil {
			return StructValueImage{}
		} else {
			return ic.EncodeObjectImage(tv.V.(Object))
		}
	case *MapType:
		if tv.V == nil {
			return MapValueImage{}
		} else {
			return ic.EncodeObjectImage(tv.V.(Object))
		}
	case blockType:
		if tv.V == nil {
			return BlockValueImage{}
		} else {
			return ic.EncodeObjectImage(tv.V.(Object))
		}
	default:
		panic("should not happen")
	}
}

func encodeObjectInfo(oi ObjectInfo) ObjectInfoImage {
	return ObjectInfoImage{
		_RealmID:      oi.ID.RealmID,
		NewTime:       oi.ID.NewTime,
		_OwnerNewTime: oi.OwnerID.NewTime,
		_ModTime:      oi.ModTime,
		_RefCount:     oi.RefCount,
	}
}

func (ic ImageCodec) DecodeTypedValueImage(tvi TypedValueImage) TypedValue {
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
