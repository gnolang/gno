package gno

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

//----------------------------------------
// (runtime) Value

type Value interface {
	assertValue()
	String() string // for debugging
}

// Fixed size primitive types are represented in TypedValue.N
// for performance.
func (StringValue) assertValue()      {}
func (BigintValue) assertValue()      {}
func (DataByteValue) assertValue()    {}
func (PointerValue) assertValue()     {}
func (*ArrayValue) assertValue()      {}
func (*SliceValue) assertValue()      {}
func (*StructValue) assertValue()     {}
func (*FuncValue) assertValue()       {}
func (*MapValue) assertValue()        {}
func (BoundMethodValue) assertValue() {}
func (TypeValue) assertValue()        {}
func (*PackageValue) assertValue()    {}
func (nativeValue) assertValue()      {}
func (blockValue) assertValue()       {}

var _ Value = StringValue("")
var _ Value = BigintValue{}
var _ Value = DataByteValue{}
var _ Value = PointerValue{}
var _ Value = &ArrayValue{} // TODO doesn't have to be pointer?
var _ Value = &SliceValue{} // TODO doesn't have to be pointer?
var _ Value = &StructValue{}
var _ Value = &FuncValue{}
var _ Value = &MapValue{}
var _ Value = BoundMethodValue{}
var _ Value = TypeValue{}
var _ Value = &PackageValue{}
var _ Value = nativeValue{}
var _ Value = blockValue{}

type StringValue string

type BigintValue struct {
	V *big.Int
}

func (bv BigintValue) Copy() BigintValue {
	return BigintValue{V: big.NewInt(0).Set(bv.V)}
}

type DataByteValue struct {
	Ref      *byte
	ElemType Type // is Uint8Kind.
}

// Base is set if the pointer refers to an array index or
// struct field or block var.
// A pointer constructed via a &x{} composite lit expression or
// constructed via new() or make() are independent objects, and
// have nil Base.
// A pointer to a block var may end up pointing to an escape
// value after a block var escapes "to the heap".
// *(PointerValue.TypedValue) must have already become
// initialized, namely T set if a typed-nil.
type PointerValue struct {
	*TypedValue        // escape val if pointer to var.
	Base        Object // array/struct/block.
	Index       int    // list/fields/values index.
}

// cu: convert untyped; pass false for const definitions
func (pv PointerValue) Assign2(rlm *Realm, tv2 TypedValue, cu bool) {
	oo1 := pv.GetObject()
	pv.Assign(tv2, cu)
	oo2 := pv.GetObject()
	rlm.DidUpdate(pv.Base, oo1, oo2)
}

func (pv PointerValue) Deref() (tv TypedValue) {
	if pv.T == DataByteType {
		dbv := pv.V.(DataByteValue)
		tv.T = dbv.ElemType
		tv.SetUint8(*dbv.Ref)
		return
	} else if nv, ok := pv.V.(*nativeValue); ok {
		rv := nv.Value
		tv.T = &nativeType{Type: rv.Type()}
		tv.V = &nativeValue{Value: rv}
		return
	} else {
		tv = *pv.TypedValue
		return
	}
}

type ArrayValue struct {
	ObjectInfo
	List []TypedValue
	Data []byte
}

func (av *ArrayValue) GetCapacity() int {
	if av.Data == nil {
		return cap(av.List)
	} else {
		return cap(av.Data)
	}
}

func (av *ArrayValue) GetLength() int {
	if av.Data == nil {
		return len(av.List)
	} else {
		return len(av.Data)
	}
}

func (av *ArrayValue) Copy() *ArrayValue {
	if av.GetRefCount() == 0 {
		return av
	}
	if av.Data == nil {
		list := make([]TypedValue, len(av.List))
		copy(list, av.List)
		return &ArrayValue{
			List: list,
		}
	} else {
		data := make([]byte, len(av.Data))
		copy(data, av.Data)
		return &ArrayValue{
			Data: data,
		}
	}
}

type SliceValue struct {
	Base   *ArrayValue
	Offset int
	Length int
	Maxcap int
}

func (sv *SliceValue) GetCapacity() int {
	return sv.Maxcap
}

func (sv *SliceValue) GetLength() int {
	return sv.Length
}

func (sv *SliceValue) GetPointerAtIndexInt2(ii int, st *SliceType) PointerValue {
	// Necessary run-time slice bounds check
	if ii < 0 {
		panic(fmt.Sprintf(
			"slice index out of bounds: %d", ii))
	} else if sv.Length <= ii {
		panic(fmt.Sprintf(
			"slice index out of bounds: %d (len=%d)",
			ii, sv.Length))
	}
	if sv.Base.Data == nil {
		ev := &sv.Base.List[sv.Offset+ii] // by reference
		if ev.IsUndefined() &&
			st.Elt.Kind() != InterfaceKind {
			// initialize typed-nil element.
			ev.T = st.Elt
			ev.V = defaultValue(st.Elt)
		}
		return PointerValue{
			TypedValue: ev,
			Base:       sv.Base,
			Index:      sv.Offset + ii,
		}
	} else {
		bv := &TypedValue{ // by reference
			T: DataByteType,
			V: DataByteValue{
				Ref:      &sv.Base.Data[sv.Offset+ii],
				ElemType: st.Elt,
			},
		}
		return PointerValue{
			TypedValue: bv,
			Base:       sv.Base,
			Index:      sv.Offset + ii,
		}
	}
}

type StructValue struct {
	ObjectInfo
	Fields []TypedValue
}

// If value is undefined at path, sets default value before
// returning.  TODO handle unexported fields in debug,
// and also ensure in the preprocessor.
func (sv *StructValue) GetPointerTo2(st *StructType, path ValuePath) PointerValue {
	if debug {
		if path.Depth != 1 {
			panic(fmt.Sprintf(
				"expected path.Depth of 1 but got %s %s",
				path.Name, path))
		}
	}
	fv := &sv.Fields[path.Index]
	if fv.IsUndefined() {
		ft := st.GetStaticTypeOfAt(path)
		if ft.Kind() == InterfaceKind {
			// Keep as undefined.
		} else {
			// Set as ft type.
			*fv = TypedValue{
				T: ft,
				V: defaultValue(ft),
			}
		}
	}
	return PointerValue{
		TypedValue: fv,
		Base:       sv,
		Index:      int(path.Index),
	}
}

func (sv *StructValue) Copy() *StructValue {
	if sv.GetRefCount() == 0 {
		return sv
	}
	fields := make([]TypedValue, len(sv.Fields))
	copy(fields, sv.Fields)
	return &StructValue{
		Fields: fields,
	}
}

// FuncValue.Type stores the method signature from the
// declaration, and has exact parameter/result names declared,
// whereas the TypedValue.T that contains at .V may not. (i.e.
// TypedValue.T doesn't care about parameter/result names, but
// the *FuncValue requires this for execution.
// In leu of FuncValue.Type, we could refer to FuncValue.Source
// or create a different field with param/result names, but
// *FuncType is already a suitable structure, and re-using
// makes construction TypedValue{T:*FuncType{},V:*FuncValue{}}
// faster.
type FuncValue struct {
	Type       *FuncType      // includes unbound receiver(s)
	IsMethod   bool           // is an (unbound) method
	Source     BlockNode      // for block mem allocation
	Name       Name           // name of function/method
	Body       []Stmt         // function body
	Closure    *Block         // creation contex (a file's Block for unbound methods).
	NativeBody func(*Machine) // alternative to Body
	FileName   Name           // file name where declared

	pkg *PackageValue
}

func (fv *FuncValue) GetPackage() *PackageValue {
	return fv.pkg
}

func (fv *FuncValue) SetPackage(pkg *PackageValue) {
	if debug {
		if fv.Type.PkgPath != pkg.PkgPath {
			panic(fmt.Sprintf(
				"function package path mismatch: %s vs %s",
				fv.Type.PkgPath,
				pkg.PkgPath))
		}
	}
	if fv.pkg != nil {
		panic("function package already set")
	}
	fv.pkg = pkg
}

type BoundMethodValue struct {
	// Underlying unbound method function.
	// The type without the receiver (since bound)
	// is computed lazily if needed.
	Func *FuncValue

	// This becomes the first arg.
	// The type is .Func.Type.Params[0].
	Receiver Value
}

type MapValue struct {
	ObjectInfo
	List *MapList

	vmap map[MapKey]*MapListItem // nil if uninitialized
}

type MapKey string

type MapList struct {
	Head *MapListItem
	Tail *MapListItem
	Size int
}

// NOTE: Value is undefined until assigned.
func (ml *MapList) Append(key TypedValue) *MapListItem {
	mli := &MapListItem{
		Prev: ml.Tail,
		Next: nil,
		Key:  key,
		// Value: undefined,
	}
	if ml.Head == nil {
		ml.Head = mli
	} else {
		// nothing
	}
	if ml.Tail != nil {
		ml.Tail.Next = mli
	}
	ml.Tail = mli
	ml.Size++
	return mli
}

func (ml *MapList) Remove(mli *MapListItem) {
	prev, next := mli.Prev, mli.Next
	if prev == nil {
		ml.Head = next
	} else {
		prev.Next = next
	}
	if next == nil {
		ml.Tail = prev
	} else {
		next.Prev = prev
	}
	ml.Size--
}

type MapListItem struct {
	Prev  *MapListItem
	Next  *MapListItem
	Key   TypedValue
	Value TypedValue
}

func (mv *MapValue) MakeMap(c int) {
	mv.List = &MapList{}
	mv.vmap = make(map[MapKey]*MapListItem, c)
}

func (mv *MapValue) GetLength() int {
	return mv.List.Size // panics if uninitialized
}

// NOTE: Go doesn't support referencing into maps, and maybe
// Gno will, but here we just use this method signature as we
// do for structs and arrays for assigning new entries.  If key
// doesn't exist, a new slot is created.
func (mv *MapValue) GetPointerForKey(key *TypedValue) PointerValue {
	kmk := key.ComputeMapKey(false)
	if mli, ok := mv.vmap[kmk]; ok {
		return PointerValue{
			TypedValue: &mli.Value,
			Base:       mv,
			Index:      -1, // no index value.
		}
	} else {
		mli := mv.List.Append(*key)
		mv.vmap[kmk] = mli
		return PointerValue{
			TypedValue: &mli.Value,
			Base:       mv,
			Index:      -1, // no index value.
		}
	}
}

// Like GetPointerForKey, but does not create a slot if key
// doesn't exist.
func (mv *MapValue) GetValueForKey(key *TypedValue) (val TypedValue, ok bool) {
	kmk := key.ComputeMapKey(false)
	if mli, exists := mv.vmap[kmk]; exists {
		val, ok = mli.Value, true
		return
	} else {
		return
	}
}

func (mv *MapValue) DeleteForKey(key *TypedValue) {
	kmk := key.ComputeMapKey(false)
	if mli, ok := mv.vmap[kmk]; ok {
		mv.List.Remove(mli)
		delete(mv.vmap, kmk)
	}
}

// The type itself as a value.
type TypeValue struct {
	Type Type
}

type PackageValue struct {
	Block
	PkgName Name
	PkgPath string
	FBlocks map[Name]*Block

	realm *Realm // if IsRealm(PkgPath)
}

func (pv *PackageValue) AddFileBlock(fn Name, b *Block) {
	if _, exists := pv.FBlocks[fn]; exists {
		panic(fmt.Sprintf(
			"duplicate file block for file %s",
			fn))
	}
	pv.FBlocks[fn] = b
}

func (pv *PackageValue) GetRealm() *Realm {
	return pv.realm
}

func (pv *PackageValue) SetRealm(rlm *Realm) {
	pv.realm = rlm
	if !pv.Block.ObjectInfo.ID.IsZero() {
		panic("should not happen")
	}
	// Set the package's ObjectInfo.ID, thereby making it real.
	pv.Block.ObjectInfo.ID = ObjectID{
		RealmID: rlm.ID,
		NewTime: 0, // 0 reserved for package block.
	}
}

type nativeValue struct {
	Value reflect.Value
}

func (nv nativeValue) Copy() nativeValue {
	nt := nv.Value.Type()
	nv2 := reflect.New(nt).Elem()
	nv2.Set(nv.Value)
	return nativeValue{Value: nv2}
}

// Only exists as PointerValue.Base.V.
type blockValue struct {
	*Block
}

//----------------------------------------
// TypedValue

type TypedValue struct {
	T Type    // never nil
	V Value   // an untyped value
	N [8]byte // numeric bytes
}

func (tv *TypedValue) IsDefined() bool {
	return !tv.IsUndefined()
}

func (tv *TypedValue) IsUndefined() bool {
	if debug {
		if tv == nil {
			panic("should not happen")
		}
	}
	if tv.T == nil {
		if debug {
			if tv.V != nil || tv.N != [8]byte{} {
				panic(fmt.Sprintf(
					"corrupted TypeValue (nil T)"))
			}
		}
		return true
	} else {
		return tv.IsNilInterface()
	}
}

func (tv *TypedValue) IsNilInterface() bool {
	if tv.T != nil && tv.T.Kind() == InterfaceKind {
		if tv.V == nil {
			return true
		} else {
			if debug {
				if tv.N != [8]byte{} {
					panic(fmt.Sprintf(
						"corrupted TypeValue (nil interface)"))
				}
			}
			return false
		}
	}
	return false
}

func (tv *TypedValue) HasKind(k Kind) bool {
	if tv.T == nil {
		return false
	} else {
		return tv.T.Kind() == k
	}
}

// for debugging, returns true if V or N is not zero.  just because V and N are
// zero doesn't mean it didn't get a value set.
func (tv *TypedValue) DebugHasValue() bool {
	if !debug {
		panic("should not happen")
	}
	if tv.V != nil {
		return true
	}
	if tv.N != [8]byte{} {
		return true
	}
	return false
}

func (tv TypedValue) String() string {
	if tv.T == nil {
		return "(undefined)"
	}
	vs := ""
	if tv.V == nil {
		switch baseOf(tv.T) {
		case BoolType, UntypedBoolType:
			vs = fmt.Sprintf("%t", tv.GetBool())
		case StringType, UntypedStringType:
			vs = fmt.Sprintf("%s", tv.GetString())
		case IntType:
			vs = fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			vs = fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			vs = fmt.Sprintf("%d", tv.GetInt16())
		case Int32Type, UntypedRuneType:
			vs = fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			vs = fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			vs = fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			vs = fmt.Sprintf("%d", tv.GetUint8())
		case DataByteType:
			vs = fmt.Sprintf("%d", tv.GetDataByte())
		case Uint16Type:
			vs = fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			vs = fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			vs = fmt.Sprintf("%d", tv.GetUint64())
		default:
			vs = "nil"
		}
	} else {
		vs = fmt.Sprintf("%v", tv.V) // reflect.TypeOf(tv.V))
	}
	ts := ""
	if tv.T == nil {
		ts = "invalid-type"
		//} else if isUntyped(tv.T) {
		//	ts = "untyped-const"
	} else {
		ts = tv.T.String()
	}
	// TODO improve.
	return fmt.Sprintf("(%s %s)", vs, ts)
}

func (tv *TypedValue) ClearNum() {
	*(*uint64)(unsafe.Pointer(&tv.N)) = uint64(0)
}

func (tv TypedValue) Copy() (cp TypedValue) {
	switch cv := tv.V.(type) {
	case BigintValue:
		cp.T = tv.T
		cp.V = cv.Copy()
	case *ArrayValue:
		cp.T = tv.T
		cp.V = cv.Copy()
	case *StructValue:
		cp.T = tv.T
		cp.V = cv.Copy()
	case nativeValue:
		cp.T = tv.T
		cp.V = cv.Copy()
	default:
		cp = tv
	}
	return
}

// Returns varint encoded bytes (and true) for numeric types
// and arbitrary bytes (and false) for variable-length byte
// types.  These bytes are used for both value hashes as well
// as hash key bytes.
func (tv *TypedValue) PrimitiveBytes() (data []byte, varint bool) {
	switch bt := baseOf(tv.T); bt {
	case BoolType:
		if tv.GetBool() {
			return []byte{0x02}, true // varint(1)
		} else {
			return []byte{0x00}, true // varint(0)
		}
	case StringType:
		return []byte(tv.GetString()), false
	case IntType:
		return varintBytes(int64(tv.GetInt())), true
	case Int8Type:
		return varintBytes(int64(tv.GetInt8())), true
	case Int16Type:
		return varintBytes(int64(tv.GetInt16())), true
	case Int32Type:
		return varintBytes(int64(tv.GetInt32())), true
	case Int64Type:
		return varintBytes(int64(tv.GetInt64())), true
	case UintType:
		return varintBytes(int64(tv.GetUint())), true
	case Uint8Type:
		return varintBytes(int64(tv.GetUint8())), true
	case Uint16Type:
		return varintBytes(int64(tv.GetUint16())), true
	case Uint32Type:
		return varintBytes(int64(tv.GetUint32())), true
	case Uint64Type:
		return varintBytes(int64(tv.GetUint64())), true
	case BigintType:
		return tv.V.(BigintValue).V.Bytes(), false
	default:
		panic(fmt.Sprintf(
			"unexpected primitive value type: %s",
			bt.String()))
	}
}

// Setting IntValue to Value is slow, and creates
// a heap allocation.  So N exists as a hack to keep
// values stored without interfaces..

func (tv *TypedValue) SetBool(b bool) {
	if debug {
		if tv.T.Kind() != BoolKind {
			panic(fmt.Sprintf(
				"TypedValue.SetBool() on type %s",
				tv.T.String()))
		}
	}
	*(*bool)(unsafe.Pointer(&tv.N)) = b
}

func (tv *TypedValue) GetBool() bool {
	if debug {
		if tv.T != nil && tv.T.Kind() != BoolKind {
			panic(fmt.Sprintf(
				"TypedValue.GetBool() on type %s",
				tv.T.String()))
		}
	}
	return *(*bool)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) GetString() string {
	if debug {
		if tv.T != nil && tv.T.Kind() != StringKind {
			panic(fmt.Sprintf(
				"TypedValue.GetString() on type %s",
				tv.T.String()))
		}
	}
	if tv.V == nil {
		return ""
	} else {
		return string(tv.V.(StringValue))
	}
}

func (tv *TypedValue) SetInt(n int) {
	if debug {
		if tv.T.Kind() != IntKind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt() on type %s",
				tv.T.String()))
		}
	}
	*(*int)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) ConvertGetInt() int {
	ConvertTo(tv, IntType)
	return tv.GetInt()
}

func (tv *TypedValue) GetInt() int {
	if debug {
		if tv.T != nil && tv.T.Kind() != IntKind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt() on type %s",
				tv.T.String()))
		}
	}
	return *(*int)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetInt8(n int8) {
	if debug {
		if tv.T.Kind() != Int8Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt8() on type %s",
				tv.T.String()))
		}
	}
	*(*int8)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetInt8() int8 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Int8Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt8() on type %s",
				tv.T.String()))
		}
	}
	return *(*int8)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetInt16(n int16) {
	if debug {
		if tv.T.Kind() != Int16Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt16() on type %s",
				tv.T.String()))
		}
	}
	*(*int16)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetInt16() int16 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Int16Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt16() on type %s",
				tv.T.String()))
		}
	}
	return *(*int16)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetInt32(n int32) {
	if debug {
		if tv.T.Kind() != Int32Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt32() on type %s",
				tv.T.String()))
		}
	}
	*(*int32)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetInt32() int32 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Int32Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt32() on type %s",
				tv.T.String()))
		}
	}
	return *(*int32)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetInt64(n int64) {
	if debug {
		if tv.T.Kind() != Int64Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt64() on type %s",
				tv.T.String()))
		}
	}
	*(*int64)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetInt64() int64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Int64Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt64() on type %s",
				tv.T.String()))
		}
	}
	return *(*int64)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetUint(n uint) {
	if debug {
		if tv.T.Kind() != UintKind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint() on type %s",
				tv.T.String()))
		}
	}
	*(*uint)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint() uint {
	if debug {
		if tv.T != nil && tv.T.Kind() != UintKind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetUint8(n uint8) {
	if debug {
		if tv.T.Kind() != Uint8Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint8() on type %s",
				tv.T.String()))
		}
		if tv.T == DataByteType {
			panic("DataByteType should call SetDataByte")
		}
	}
	*(*uint8)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint8() uint8 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Uint8Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint8() on type %s",
				tv.T.String()))
		}
		if tv.T == DataByteType {
			panic("DataByteType should call GetDataByte or GetUint8OrDataByte")
		}
	}
	return *(*uint8)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetDataByte(n uint8) {
	if debug {
		if tv.T != DataByteType {
			panic(fmt.Sprintf(
				"TypedValue.SetDataByte() on type %s",
				tv.T.String()))
		}
	}
	*(tv.V.(DataByteValue).Ref) = n
}

func (tv *TypedValue) GetDataByte() uint8 {
	if debug {
		if tv.T != nil && tv.T != DataByteType {
			panic(fmt.Sprintf(
				"TypedValue.GetDataByte() on type %s",
				tv.T.String()))
		}
	}
	return *(tv.V.(DataByteValue).Ref)
}

func (tv *TypedValue) SetUint16(n uint16) {
	if debug {
		if tv.T.Kind() != Uint16Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint16() on type %s",
				tv.T.String()))
		}
	}
	*(*uint16)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint16() uint16 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Uint16Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint16() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint16)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetUint32(n uint32) {
	if debug {
		if tv.T.Kind() != Uint32Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint32() on type %s",
				tv.T.String()))
		}
	}
	*(*uint32)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint32() uint32 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Uint32Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint32() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint32)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetUint64(n uint64) {
	if debug {
		if tv.T.Kind() != Uint64Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint64() on type %s",
				tv.T.String()))
		}
	}
	*(*uint64)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint64() uint64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Uint64Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint64() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint64)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) GetBig() *big.Int {
	if debug {
		if tv.T != nil && tv.T.Kind() != BigintKind {
			panic(fmt.Sprintf(
				"TypedValue.GetBig() on type %s",
				tv.T.String()))
		}
	}
	return tv.V.(BigintValue).V
}

func (tv *TypedValue) ComputeMapKey(omitType bool) MapKey {
	// Special case when nil: has no separator.
	if tv.T == nil {
		if debug {
			if omitType {
				panic("should not happen")
			}
		}
		return MapKey("nil")
	}
	// General case.
	bz := make([]byte, 0, 64)
	if !omitType {
		bz = append(bz, tv.T.TypeID().Bytes()...)
		bz = append(bz, ':') // type/value separator
	}
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		pbz, _ := tv.PrimitiveBytes()
		bz = append(bz, pbz...)
	case PointerType:
		ptr := uintptr(unsafe.Pointer(tv.V.(PointerValue).TypedValue))
		bz = append(bz, uintptrToBytes(&ptr)...)
	case FieldType:
		panic("field (pseudo)type cannot be used as map key")
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		al := av.GetLength()
		bz = append(bz, '[')
		if av.Data == nil {
			omitTypes := bt.Elem().Kind() != InterfaceKind
			for i := 0; i < al; i++ {
				ev := &av.List[i]
				bz = append(bz, ev.ComputeMapKey(omitTypes)...)
				if i != al-1 {
					bz = append(bz, ',')
				}
			}
		} else {
			bz = append(bz, av.Data...)
		}
		bz = append(bz, ']')
	case *SliceType:
		panic("slice type cannot be used as map key")
	case *StructType:
		sv := tv.V.(*StructValue)
		sl := len(sv.Fields)
		bz = append(bz, '{')
		for i := 0; i < sl; i++ {
			fv := &sv.Fields[i]
			ft := bt.Fields[i]
			omitTypes := ft.Elem().Kind() != InterfaceKind
			bz = append(bz, fv.ComputeMapKey(omitTypes)...)
			if i != sl-1 {
				bz = append(bz, ',')
			}
		}
		bz = append(bz, '}')
	case *FuncType:
		panic("func type cannot be used as map key")
	case *MapType:
		panic("map type cannot be used as map key")
	case *InterfaceType:
		panic("should not happen")
	case *PackageType:
		pv := tv.V.(*PackageValue)
		bz = append(bz, []byte(strconv.Quote(pv.PkgPath))...)
	case *ChanType:
		panic("not yet implemented")
	case *nativeType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"unexpected map key type %s",
			tv.T.String()))
	}
	return MapKey(bz)
}

//----------------------------------------
// Value utility/manipulation functions.

// cu: convert untyped after assignment. pass false
// for const definitions, but true for all else.
func (tv *TypedValue) Assign(tv2 TypedValue, cu bool) {
	if debug {
		if tv2.T == DataByteType {
			// tv2 will never be a DataByte, as it is
			// retrieved as value.
			panic("should not happen")
		}
	}
	switch ct := baseOf(tv.T).(type) {
	case PrimitiveType:
		if ct == DataByteType {
			tv.SetDataByte(tv2.GetUint8())
		} else {
			*tv = tv2.Copy()
			if cu && isUntyped(tv.T) {
				ConvertUntypedTo(tv, defaultTypeOf(tv.T))
			}
		}
	case *nativeType:
		switch v2 := tv2.V.(type) {
		case PointerValue:
			nv1 := tv.V.(*nativeValue)
			if ct.Type.Kind() != reflect.Ptr {
				panic("should not happen")
			}
			if nv2, ok := v2.TypedValue.V.(*nativeValue); ok {
				nrv2 := nv2.Value
				if nrv2.CanAddr() {
					it := nrv2.Addr()
					nv1.Value.Set(it)
				} else {
					// XXX think more
					panic("not yet implemented")
				}
			} else {
				// XXX think more
				panic("not yet implemented")
			}
		case *nativeValue:
			if tv.V == nil {
				if debug {
					// tv.V is a native function type.
					if tv.T.Kind() != FuncKind ||
						tv2.T.Kind() != FuncKind {
						panic("should not happen")
					}
					if nv, ok := tv2.V.(*nativeValue); !ok ||
						nv.Value.Kind() != reflect.Func {
						panic("should not happen")
					}
				}
				tv.V = v2
			} else {
				nv1 := tv.V.(*nativeValue)
				nv1.Value.Set(v2.Value)
			}
		default:
			panic("should not happen")
		}
	case *StructType: // XXX path index+other.
		if tv2.IsUndefined() {
			*tv = TypedValue{}
		} else {
			if debug {
				if tv.T.TypeID() != tv2.T.TypeID() {
					panic(fmt.Sprintf("mismatched types: cannot assign %v to %v",
						tv2.String(), tv.T.String()))
				}
			}
			// XXX fix realm refcount logic.
			*tv = tv2.Copy()
		}
	case nil:
		*tv = tv2.Copy()
		if cu && isUntyped(tv.T) {
			ConvertUntypedTo(tv, defaultTypeOf(tv.T))
		} else {
			// pass cu=false for const definitions.
		}
	default:
		*tv = tv2.Copy()
	}
}

func (tv *TypedValue) ConvertUntyped() {
	if isUntyped(tv.T) {
		ConvertUntypedTo(tv, defaultTypeOf(tv.T))
	}
}

func (tv *TypedValue) GetPointerTo(path ValuePath) PointerValue {
	if debug {
		if tv.IsUndefined() {
			panic("GetPointerTo() on undefined value")
		}
	}
	t, v := tv.T, tv.V
TYPE_SWITCH:
	switch ct := t.(type) {
	case *DeclaredType:
		if path.Depth <= 1 {
			switch path.Type {
			case VPTypeDefault:
				/*
					main.Bir struct{(struct{("foo" string)} main.Boo)} @Hello
					PATH=gno.ValuePath{Type:0x2, Depth:0x1, Index:0x0, Name:"Hello"}
				*/
				ftv := ct.GetValueRefAt(path)
				fv := ftv.GetFunc()
				mv := BoundMethodValue{
					Func:     fv,
					Receiver: tv.V, // use original v
				}
				// TODO: this means all method selectors are slow
				// and incur extra overhead.  To prevent this extra
				// overhead, CallExpr evaluation should keep into X
				// and call *DeclaredType.GetMethodAt() directly.
				return PointerValue{
					TypedValue: &TypedValue{
						T: fv.Type.BoundType(),
						V: mv,
					},
					Base: nil, // a bound method is free floating.
				}
			case VPTypeInterface:
				tr, _, _, _ := ct.FindEmbeddedFieldType(path.Name)
				if len(tr) == 0 {
					panic("should not happen")
				}
				bv := tv
				for i, path := range tr {
					ptr := bv.GetPointerTo(path)
					if i == len(tr)-1 {
						return ptr // done
					} else {
						bv = ptr.TypedValue // deref
					}
				}
				panic("should not happen")
			default:
				panic("should not happen")
			}
		} else {
			// NOTE could work with nested *DeclaredTypes,
			// though we don't yet allow that.
			path.Depth--
			t = ct.Base
			goto TYPE_SWITCH
		}
	case PointerType:
		switch cet := ct.Elt.(type) {
		case *DeclaredType:
			t = cet
			v = v.(PointerValue).V
			goto TYPE_SWITCH
		case *nativeType:
			t = cet
			v = v.(PointerValue).V
			goto TYPE_SWITCH
		default:
			panic(fmt.Sprintf(
				"unexpected pointer type: %v",
				ct.String()))
		}
	case *StructType:
		return v.(*StructValue).GetPointerTo2(ct, path)
	case *TypeType:
		switch t := v.(TypeValue).Type.(type) {
		case *DeclaredType:
			return PointerValue{
				TypedValue: t.GetValueRefAt(path),
				Base:       nil, // TODO: make TypeValue an object.
			}
		case *nativeType:
			rt := t.Type
			mt, ok := rt.MethodByName(string(path.Name))
			if !ok {
				if debug {
					panic(fmt.Sprintf(
						"native type %s has no method %s",
						rt.String(), path.Name))
				}
				panic("unknown native method selector")
			}
			mtv := go2GnoValue(mt.Func)
			return PointerValue{
				TypedValue: &mtv, // heap alloc
				Base:       nil,
			}
		default:
			panic("unexpected selector base typeval.")
		}
	case *PackageType:
		pv := v.(*PackageValue)
		return pv.GetPointerTo(path)
	case *nativeType:
		// Special case if tv.T.(PointerType):
		// we may need to treat this as a native pointer
		// to get the correct pointer-receiver value.
		var rv reflect.Value
		if pv, isGnoPtr := v.(PointerValue); isGnoPtr {
			rv = pv.TypedValue.V.(*nativeValue).Value
		} else {
			rv = v.(*nativeValue).Value
		}
		rt := rv.Type()
		// First, try to get field.
		var fv reflect.Value
		if rt.Kind() == reflect.Ptr {
			if rt.Elem().Kind() == reflect.Struct {
				fv = rv.Elem().FieldByName(string(path.Name))
			}
		} else if rt.Kind() == reflect.Struct {
			fv = rv.FieldByName(string(path.Name))
		}
		if fv.IsValid() {
			return PointerValue{
				TypedValue: &TypedValue{ // heap alloc
					T: &nativeType{Type: fv.Type()},
					V: &nativeValue{Value: fv},
				},
				Base: nil,
			}
		}
		// Then, try to get method.
		mv := rv.MethodByName(string(path.Name))
		if mv.IsValid() {
			mt := mv.Type()
			return PointerValue{
				TypedValue: &TypedValue{ // heap alloc
					T: &nativeType{Type: mt},
					V: &nativeValue{Value: mv},
				},
				Base: nil,
			}
		} else {
			// If isGnoPtr, try to get method from pointer type.
			if !rv.CanAddr() {
				// Replace rv with addressable value.
				rv2 := reflect.New(rt).Elem()
				rv2.Set(rv)
				rv = rv2
				tv.V.(*nativeValue).Value = rv2 // replace rv
			}
			mv := rv.Addr().MethodByName(string(path.Name))
			if mv.IsValid() {
				mt := mv.Type()
				return PointerValue{
					TypedValue: &TypedValue{ // heap alloc
						T: &nativeType{Type: mt},
						V: &nativeValue{Value: mv},
					},
					Base: nil,
				}
			}

		}
		panic(fmt.Sprintf(
			"native type %s has no method or field %s",
			ct.String(), path.Name))
	default:
		panic(fmt.Sprintf(
			"unexpected selector base type for mutation: %s.",
			t.String()))
	}
}

// Convenience for GetPointerAtIndex().  Slow.
func (tv *TypedValue) GetPointerAtIndexInt(ii int) PointerValue {
	iv := TypedValue{T: IntType}
	iv.SetInt(ii)
	return tv.GetPointerAtIndex(&iv)
}

// If element value is undefined and the array/slice is not of
// interfaces, the appropriate type is first set.
func (tv *TypedValue) GetPointerAtIndex(iv *TypedValue) PointerValue {
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		if bt == StringType {
			sv := tv.GetString()
			ii := iv.ConvertGetInt()
			bv := &TypedValue{ // heap alloc
				T: Uint8Type,
			}
			bv.SetUint8(sv[ii])
			return PointerValue{
				TypedValue: bv,
				Base:       nil, // free floating
			}
		} else {
			panic(fmt.Sprintf(
				"primitive type %s cannot be indexed",
				tv.T.String()))
		}
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		ii := iv.ConvertGetInt()
		if av.Data == nil {
			ev := &av.List[ii] // by reference
			if ev.IsUndefined() &&
				bt.Elt.Kind() != InterfaceKind {
				// initialize typed-nil element.
				ev.T = bt.Elt
				ev.V = defaultValue(bt.Elt)
			}
			return PointerValue{
				TypedValue: ev,
				Base:       av,
				Index:      ii,
			}
		} else {
			bv := &TypedValue{ // heap alloc
				T: DataByteType,
				V: DataByteValue{
					Ref:      &av.Data[ii],
					ElemType: bt.Elem(),
				},
			}
			return PointerValue{
				TypedValue: bv,
				Base:       av,
				Index:      ii,
			}
		}
	case *SliceType:
		if tv.V == nil {
			panic("nil slice index (out of bounds)")
		}
		sv := tv.V.(*SliceValue)
		ii := iv.ConvertGetInt()
		return sv.GetPointerAtIndexInt2(ii, bt)
	case *MapType:
		if tv.V == nil {
			panic("uninitialized map index")
		}
		mv := tv.V.(*MapValue)
		pv := mv.GetPointerForKey(iv)
		// XXX implement x, ok := m[idx]
		if pv.TypedValue.IsUndefined() {
			vt := baseOf(tv.T).(*MapType).Value
			if vt.Kind() != InterfaceKind {
				// initialize typed-nil key.
				pv.TypedValue.T = vt
			}
		}
		return pv
	case *nativeType:
		rv := tv.V.(*nativeValue).Value
		ii := iv.ConvertGetInt()
		ev := rv.Index(ii)
		etv := go2GnoValue(ev)
		return PointerValue{
			TypedValue: &etv,
			Base:       nil,
		}
	default:
		panic(fmt.Sprintf(
			"unexpected index base type %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetType() Type {
	return tv.V.(TypeValue).Type
}

func (tv *TypedValue) GetFunc() *FuncValue {
	return tv.V.(*FuncValue)
}

func (tv *TypedValue) GetLength() int {
	if tv.V == nil {
		switch bt := baseOf(tv.T).(type) {
		case PrimitiveType:
			if bt != StringType {
				panic("should not happen")
			}
			return 0
		case *ArrayType:
			return bt.Len
		case *SliceType:
			return 0
		default:
			panic(fmt.Sprintf(
				"unexpected type for len(): %s",
				bt.String()))
		}
	}
	switch cv := tv.V.(type) {
	case StringValue:
		return len(cv)
	case *ArrayValue:
		return cv.GetLength()
	case *SliceValue:
		return cv.GetLength()
	case *MapValue:
		return cv.GetLength()
	default:
		panic(fmt.Sprintf("unexpected type for len(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetCapacity() int {
	if tv.V == nil {
		if debug {
			// assert acceptable type.
			switch baseOf(tv.T).(type) {
			// strings have no capacity.
			case *ArrayType:
			case *SliceType:
			default:
				panic("should not happen")
			}
		}
		return 0
	}
	switch cv := tv.V.(type) {
	case *ArrayValue:
		return cv.GetCapacity()
	case *SliceValue:
		return cv.GetCapacity()
	default:
		panic(fmt.Sprintf("unexpected type for cap(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetSlice(low, high int) TypedValue {
	if low < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			low))
	}
	if high < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			high))
	}
	if low > high {
		panic(fmt.Sprintf(
			"invalid slice index %d > %d",
			low, high))
	}
	if tv.GetCapacity() < high {
		panic(fmt.Sprintf(
			"slice bounds out of range [%d:%d] with capacity %d",
			low, high, tv.GetCapacity()))
	}
	switch t := baseOf(tv.T).(type) {
	case PrimitiveType:
		if t == StringType {
			return TypedValue{
				T: tv.T,
				V: StringValue(tv.GetString()[low:high]),
			}
		} else {
			panic("non-string primitive type cannot be sliced")
		}
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		st := &SliceType{
			Elt: t.Elt,
			Vrd: false,
		}
		return TypedValue{
			T: st,
			V: &SliceValue{
				Base:   av,
				Offset: low,
				Length: high - low,
				Maxcap: av.GetCapacity() - low,
			},
		}
	case *SliceType:
		if tv.V == nil {
			if low != 0 || high != 0 {
				panic("nil slice index out of range")
			}
			return TypedValue{
				T: tv.T,
				V: nil,
			}
		}
		sv := tv.V.(*SliceValue)
		return TypedValue{
			T: tv.T,
			V: &SliceValue{
				Base:   sv.Base,
				Offset: sv.Offset + low,
				Length: sv.Offset + high - low,
				Maxcap: sv.Maxcap - low,
			},
		}
	default:
		panic(fmt.Sprintf("unexpected type for GetSlice(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetSlice2(low, high, max int) TypedValue {
	if low < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			low))
	}
	if high < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			high))
	}
	if max < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			max))
	}
	if low > high {
		panic(fmt.Sprintf(
			"invalid slice index %d > %d",
			low, high))
	}
	if high > max {
		panic(fmt.Sprintf(
			"invalid slice index %d > %d",
			high, max))
	}
	if tv.GetCapacity() < high {
		panic(fmt.Sprintf(
			"slice bounds out of range [%d:%d:%d] with capacity %d",
			low, high, max, tv.GetCapacity()))
	}
	if tv.GetCapacity() < max {
		panic(fmt.Sprintf(
			"slice bounds out of range [%d:%d:%d] with capacity %d",
			low, high, max, tv.GetCapacity()))
	}
	switch bt := baseOf(tv.T).(type) {
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		st := &SliceType{
			Elt: bt.Elt,
			Vrd: false,
		}
		return TypedValue{
			T: st,
			V: &SliceValue{
				Base:   av,
				Offset: low,
				Length: high - low,
				Maxcap: max - low,
			},
		}
	case *SliceType:
		if tv.V == nil {
			if low != 0 || high != 0 || max != 0 {
				panic("nil slice index out of range")
			}
			return TypedValue{
				T: tv.T,
				V: nil,
			}
		}
		sv := tv.V.(*SliceValue)
		return TypedValue{
			T: tv.T,
			V: &SliceValue{
				Base:   sv.Base,
				Offset: sv.Offset + low,
				Length: sv.Offset + high - low,
				Maxcap: max - low,
			},
		}
	default:
		panic(fmt.Sprintf("unexpected type for GetSlice2(): %s",
			tv.T.String()))
	}
}

// Returns the field type of container type.
// This is distinct from tv.GetValueRefAt(path).T for:
// * uninitialized struct, package fields
// * interface fields
func (tv *TypedValue) GetStaticTypeOfAt(path ValuePath) Type {
	t := tv.T
TYPE_SWITCH:
	switch ct := t.(type) {
	case *DeclaredType:
		if path.Depth <= 1 {
			ftv := ct.GetValueRefAt(path)
			return ftv.T.(*FuncType).BoundType()
		} else {
			path.Depth--
			t = ct.Base
			goto TYPE_SWITCH
		}
	case *StructType:
		return ct.GetStaticTypeOfAt(path)
	case *PackageType:
		return tv.V.(*PackageValue).Source.GetStaticTypeOfAt(path)
	default:
		panic("should not happen")
	}
}

//----------------------------------------
// Block
//
// Blocks hold values referred to by var/const/func/type
// declarations in BlockNodes such as packages, functions,
// and switch statements.  Unlike structs or packages,
// names and paths may refer to parent blocks.  (In the
// future, the same mechanism may be used to support
// inheritance or prototype-like functionality for structs
// and packages.)
//
// When a block would otherwise become gc'd because it is no
// longer used except for escaped reference pointers to
// variables, and there are no closures that reference the
// block, the remaining references to objects become detached
// from the block and become ownerless.

type Block struct {
	ObjectInfo // for closures
	Source     BlockNode
	Values     []TypedValue
	Parent     *Block
	Blank      TypedValue // captures "_"
	bodyStmt   bodyStmt
}

func NewBlock(source BlockNode, parent *Block) *Block {
	var values []TypedValue
	if source != nil {
		values = make([]TypedValue, source.GetNumNames())
	}
	return &Block{
		Source: source,
		Values: values,
		Parent: parent,
	}
}

func (b *Block) String() string {
	return b.StringIndented("    ")
}

func (b *Block) StringIndented(indent string) string {
	source := toString(b.Source)
	if len(source) > 16 {
		source = source[:14] + "..."
	}
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("Block(Addr:%p,Source:%s,Parent:%p)",
			b, source, b.Parent))
	if b.Source != nil {
		for i, n := range b.Source.GetNames() {
			if len(b.Values) <= i {
				lines = append(lines,
					fmt.Sprintf("%s%s: undefined", indent, n))
			} else {
				lines = append(lines,
					fmt.Sprintf("%s%s: %s",
						indent, n, b.Values[i].String()))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (b *Block) GetPointerTo(path ValuePath) PointerValue {
	if path.IsZeroPath() {
		if debug {
			if path.Name != "_" {
				panic(fmt.Sprintf(
					"zero value path is reserved for \"_\", but got %s",
					path.Name))
			}
		}
		return PointerValue{
			TypedValue: b.GetBlankRef(),
			Base:       b,
			Index:      -1,
		}
	}
	// NOTE: For most block paths, Depth starts at 1, but
	// the generation for uverse is 0.  If path.Depth is
	// 0, it implies that b == uverse, and the condition
	// would fail as if it were 1.
	i := uint8(1)
LOOP:
	if i < path.Depth {
		b = b.Parent
		i++
		goto LOOP
	}
	return PointerValue{
		TypedValue: &b.Values[path.Index],
		Base:       b,
		Index:      int(path.Index),
	}
}

// Result is used has lhs for any assignments to "_".
func (b *Block) GetBlankRef() *TypedValue {
	return &b.Blank
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams1() (pv1 PointerValue) {
	pv1 = b.GetPointerTo(NewValuePathDefault(1, 0, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams2() (pv1, pv2 PointerValue) {
	pv1 = b.GetPointerTo(NewValuePathDefault(1, 0, ""))
	pv2 = b.GetPointerTo(NewValuePathDefault(1, 1, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams3() (pv1, pv2, pv3 PointerValue) {
	pv1 = b.GetPointerTo(NewValuePathDefault(1, 0, ""))
	pv2 = b.GetPointerTo(NewValuePathDefault(1, 1, ""))
	pv3 = b.GetPointerTo(NewValuePathDefault(1, 2, ""))
	return
}

func (b *Block) GetBodyStmt() *bodyStmt {
	return &b.bodyStmt
}

// Used by SwitchStmt upon clause match.
func (b *Block) ExpandToSize(size uint16) {
	if debug {
		if len(b.Values) >= int(size) {
			panic("unexpected block size shrinkage")
		}
	}
	values := make([]TypedValue, int(size))
	copy(values, b.Values)
	b.Values = values
}

//----------------------------------------

func defaultValue(t Type) Value {
	switch ct := baseOf(t).(type) {
	case *ArrayType:
		tvs := make([]TypedValue, ct.Len)
		return &ArrayValue{
			List: tvs,
		}
	case *SliceType:
		return &SliceValue{
			Base: nil,
		}
	case *MapType:
		// zero uninitialized maps are not valid.
		panic("should not happen")
	case *StructType:
		return &StructValue{
			Fields: make([]TypedValue, len(ct.Fields)),
		}
	case *nativeType:
		return &nativeValue{
			Value: reflect.New(ct.Type).Elem(),
		}
	default:
		return nil
	}
}

func untypedBool(b bool) TypedValue {
	tv := TypedValue{T: UntypedBoolType}
	tv.SetBool(b)
	return tv
}

func typedRune(r rune) TypedValue {
	tv := TypedValue{T: Int32Type}
	tv.SetInt32(int32(r))
	return tv
}

func newSliceFromList(list []TypedValue) *SliceValue {
	return &SliceValue{
		Base: &ArrayValue{
			List: list,
		},
		Offset: 0,
		Length: len(list),
		Maxcap: cap(list),
	}
}

func newSliceFromData(data []byte) *SliceValue {
	return &SliceValue{
		Base: &ArrayValue{
			Data: data,
		},
		Offset: 0,
		Length: len(data),
		Maxcap: cap(data),
	}
}
