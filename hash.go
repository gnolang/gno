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
	String() string
}

type ObjectInfoImage struct {
	ID       ObjectID
	OwnerID  ObjectID
	ModTime  uint64
	RefCount int
}

// XXX DataByte:
// XXX is a reference to ArrayValue.Data[i],
// XXX in place of what would usually be a Uint8Kind.
// XXX this works at runtime, but at persist time,
// XXX we would need to know the source/base ArrayValue
// XXX and index.  So, how?

type NumValueImage [8]byte

type StringValueImage string

type PointerValueImage struct {
	BaseID          ObjectID // if weak
	Index           int      // if weak
	TypedValueImage          // if owned
}

type ArrayValueImage struct {
	ObjectInfo ObjectInfoImage
	List       []TypedValueImage
	Data       []byte
}

type SliceValueImage struct {
	BaseID          ObjectID // if weak
	ArrayValueImage          // if owned
	Offset          int
	Length          int
	Maxcap          int
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
	Values     []TypedValueImage
	ParentID   ObjectID
}

//----------------------------------------

type TypeLookup func(TypeID) Type

func EncodeTypedValueImage(tv TypedValue) TypedValueImage {
	return TypedValueImage{}
}

func DecodeTypedValueImage(tl TypeLookup, tvi TypedValueImage) TypedValue {
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
