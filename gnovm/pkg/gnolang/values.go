package gnolang

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/cockroachdb/apd/v3"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// ----------------------------------------
// (runtime) Value

type Value interface {
	assertValue()
	String() string // for debugging
}

// Fixed size primitive types are represented in TypedValue.N
// for performance.
func (StringValue) assertValue()       {}
func (BigintValue) assertValue()       {}
func (BigdecValue) assertValue()       {}
func (DataByteValue) assertValue()     {}
func (PointerValue) assertValue()      {}
func (*ArrayValue) assertValue()       {}
func (*SliceValue) assertValue()       {}
func (*StructValue) assertValue()      {}
func (*FuncValue) assertValue()        {}
func (*MapValue) assertValue()         {}
func (*BoundMethodValue) assertValue() {}
func (TypeValue) assertValue()         {}
func (*PackageValue) assertValue()     {}
func (*NativeValue) assertValue()      {}
func (*Block) assertValue()            {}
func (RefValue) assertValue()          {}

const (
	nilStr       = "nil"
	undefinedStr = "undefined"
)

var (
	_ Value = StringValue("")
	_ Value = BigintValue{}
	_ Value = BigdecValue{}
	_ Value = DataByteValue{}
	_ Value = PointerValue{}
	_ Value = &ArrayValue{} // TODO doesn't have to be pointer?
	_ Value = &SliceValue{} // TODO doesn't have to be pointer?
	_ Value = &StructValue{}
	_ Value = &FuncValue{}
	_ Value = &MapValue{}
	_ Value = &BoundMethodValue{}
	_ Value = TypeValue{}
	_ Value = &PackageValue{}
	_ Value = &NativeValue{}
	_ Value = &Block{}
	_ Value = RefValue{}
)

// ----------------------------------------
// StringValue

type StringValue string

// ----------------------------------------
// BigintValue

type BigintValue struct {
	V *big.Int
}

func (bv BigintValue) MarshalAmino() (string, error) {
	bz, err := bv.V.MarshalText()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (bv *BigintValue) UnmarshalAmino(s string) error {
	vv := big.NewInt(0)
	err := vv.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	bv.V = vv
	return nil
}

func (bv BigintValue) Copy(alloc *Allocator) BigintValue {
	return BigintValue{V: big.NewInt(0).Set(bv.V)}
}

// ----------------------------------------
// BigdecValue

type BigdecValue struct {
	V *apd.Decimal
}

func (bv BigdecValue) MarshalAmino() (string, error) {
	bz, err := bv.V.MarshalText()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (bv *BigdecValue) UnmarshalAmino(s string) error {
	vv := apd.New(0, 0)
	err := vv.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	bv.V = vv
	return nil
}

func (bv BigdecValue) Copy(alloc *Allocator) BigdecValue {
	cp := apd.New(0, 0)
	_, err := apd.BaseContext.Add(cp, cp, bv.V)
	if err != nil {
		panic("should not happen")
	}
	return BigdecValue{V: cp}
}

// ----------------------------------------
// DataByteValue

type DataByteValue struct {
	Base     *ArrayValue // base array.
	Index    int         // base.Data index.
	ElemType Type        // is Uint8Kind.
}

func (dbv DataByteValue) GetByte() byte {
	return dbv.Base.Data[dbv.Index]
}

func (dbv DataByteValue) SetByte(b byte) {
	dbv.Base.Data[dbv.Index] = b
}

// ----------------------------------------
// PointerValue

// Base is set if the pointer refers to an array index or
// struct field or block var.
// A pointer constructed via a &x{} composite lit
// expression or constructed via new() or make() are
// independent objects, and have nil Base.
// A pointer to a block var may end up pointing to an escape
// value after a block var escapes "to the heap".
// *(PointerValue.TypedValue) must have already become
// initialized, namely T set if a typed-nil.
// Index is -1 for the shared "_" block var,
// and -2 for (gno and native) map items.
//
// Allocation for PointerValue is not immediate,
// as usually PointerValues are temporary for assignment
// or binary operations. When a pointer is to be
// allocated, *Allocator.AllocatePointer() is called separately,
// as in OpRef.
type PointerValue struct {
	TV    *TypedValue // escape val if pointer to var.
	Base  Value       // array/struct/block.
	Index int         // list/fields/values index, or -1 or -2 (see below).
	Key   *TypedValue `json:",omitempty"` // for maps.
}

const (
	PointerIndexBlockBlank = -1 // for the "_" identifier in blocks
	PointerIndexMap        = -2 // Base is Map, use Key.
	PointerIndexNative     = -3 // Base is *NativeValue.
)

/*
func (pv *PointerValue) GetBase(store Store) Object {
	switch cbase := pv.Base.(type) {
	case nil:
		return nil
	case RefValue:
		base := store.GetObject(cbase.ObjectID).(Object)
		pv.Base = base
		return base
	case Object:
		return cbase
	default:
		panic("should not happen")
	}
}
*/

// cu: convert untyped; pass false for const definitions
// TODO: document as something that enables into-native assignment.
// TODO: maybe consider this as entrypoint for DataByteValue too?
func (pv PointerValue) Assign2(alloc *Allocator, store Store, rlm *Realm, tv2 TypedValue, cu bool) {
	// Special cases.
	if pv.Index == PointerIndexNative {
		// Special case if extended object && native.
		rv := pv.Base.(*NativeValue).Value
		if rv.Kind() == reflect.Map { // go native object
			// assign value to map directly.
			krv := gno2GoValue(pv.Key, reflect.Value{})
			vrv := gno2GoValue(&tv2, reflect.Value{})
			rv.SetMapIndex(krv, vrv)
		} else {
			// assign depending on pv.TV type.
			tv := pv.TV
			nv1 := tv.V.(*NativeValue)
			switch v2 := tv2.V.(type) {
			case PointerValue:
				if tv.T.Kind() != PointerKind {
					panic("should not happen")
				}
				if nv2, ok := v2.TV.V.(*NativeValue); ok {
					nrv2 := nv2.Value
					if nrv2.CanAddr() {
						it := nrv2.Addr()
						nv1.Value.Set(it)
					} else {
						panic("not yet implemented")
					}
				} else {
					panic("not yet implemented")
				}
			case *NativeValue:
				if tv.V == nil {
					// tv.V is a native function type.
					// there is no default value, so just assign
					// rather than .Value.Set().
					if tv.T.Kind() == FuncKind {
						if debug {
							if tv2.T.Kind() != FuncKind {
								panic("should not happen")
							}
							if nv, ok := tv2.V.(*NativeValue); !ok ||
								nv.Value.Kind() != reflect.Func {
								panic("should not happen")
							}
						}
						tv.V = v2
					} else {
						tv.V = defaultValue(alloc, tv.T)
						tv.V.(*NativeValue).Value.Set(v2.Value)
					}
				} else {
					nv1.Value.Set(v2.Value)
				}
			case nil:
				if debug {
					if tv2.T != nil && tv.T.TypeID() != tv2.T.TypeID() {
						panic(fmt.Sprintf("mismatched types: cannot assign %v to %v",
							tv2.String(), tv.T.String()))
					}
				}
				*tv = tv2.Copy(alloc)
			default:
				panic("should not happen")
			}
		}
		return
	} else if pv.TV.T == DataByteType {
		// Special case of DataByte into (base=*SliceValue).Data.
		pv.TV.SetDataByte(tv2.GetUint8())
		return
	}
	// General case
	if rlm != nil && pv.Base != nil {
		oo1 := pv.TV.GetFirstObject(store)
		pv.TV.Assign(alloc, tv2, cu)
		oo2 := pv.TV.GetFirstObject(store)
		rlm.DidUpdate(pv.Base.(Object), oo1, oo2)
	} else {
		pv.TV.Assign(alloc, tv2, cu)
	}
}

func (pv PointerValue) Deref() (tv TypedValue) {
	if pv.TV.T == DataByteType {
		dbv := pv.TV.V.(DataByteValue)
		tv.T = dbv.ElemType
		tv.SetUint8(dbv.GetByte())
		return
	} else if nv, ok := pv.TV.V.(*NativeValue); ok {
		rv := nv.Value
		// XXX memoize type.
		tv.T = &NativeType{Type: rv.Type()}
		tv.V = nv
		return
	} else {
		tv = *pv.TV
		return
	}
}

// ----------------------------------------
// ArrayValue

type ArrayValue struct {
	ObjectInfo
	List []TypedValue
	Data []byte
}

// NOTE: Result should not be written to,
// behavior is unexpected when .List bytes.
func (av *ArrayValue) GetReadonlyBytes() []byte {
	if av.Data == nil {
		// NOTE: we cannot convert to .Data type bytearray here
		// because there might be references to .List[x].
		bz := make([]byte, len(av.List))
		for i, tv := range av.List {
			if tv.T.Kind() != Uint8Kind {
				panic(fmt.Sprintf(
					"expected byte kind but got %v",
					tv.T.Kind()))
			}
			bz[i] = tv.GetUint8()
		}
		return bz
	}
	return av.Data
}

func (av *ArrayValue) GetCapacity() int {
	if av.Data == nil {
		// not cap(av.List) for simplicity.
		// extra capacity is ignored.
		return len(av.List)
	}
	// not cap(av.Data) for simplicity.
	// extra capacity is ignored.
	return len(av.Data)
}

func (av *ArrayValue) GetLength() int {
	if av.Data == nil {
		return len(av.List)
	}
	return len(av.Data)
}

// et is only required for .List byte-arrays.
func (av *ArrayValue) GetPointerAtIndexInt2(store Store, ii int, et Type) PointerValue {
	if av.Data == nil {
		ev := fillValueTV(store, &av.List[ii]) // by reference
		return PointerValue{
			TV:    ev,
			Base:  av,
			Index: ii,
		}
	}
	bv := &TypedValue{ // heap alloc, so need to compare value rather than pointer
		T: DataByteType,
		V: DataByteValue{
			Base:     av,
			Index:    ii,
			ElemType: et,
		},
	}

	return PointerValue{
		TV:    bv,
		Base:  av,
		Index: ii,
	}
}

func (av *ArrayValue) Copy(alloc *Allocator) *ArrayValue {
	/* TODO: consider second ref count field.
	if av.GetRefCount() == 0 {
		return av
	}
	*/
	if av.Data == nil {
		av2 := alloc.NewListArray(len(av.List))
		copy(av2.List, av.List)
		return av2
	}
	av2 := alloc.NewDataArray(len(av.Data))
	copy(av2.Data, av.Data)
	return av2
}

// ----------------------------------------
// SliceValue

type SliceValue struct {
	Base   Value
	Offset int
	Length int
	Maxcap int
}

func (sv *SliceValue) GetBase(store Store) *ArrayValue {
	switch cv := sv.Base.(type) {
	case nil:
		return nil
	case RefValue:
		array := store.GetObject(cv.ObjectID).(*ArrayValue)
		sv.Base = array
		return array
	case *ArrayValue:
		return cv
	default:
		panic("should not happen")
	}
}

func (sv *SliceValue) GetCapacity() int {
	return sv.Maxcap
}

func (sv *SliceValue) GetLength() int {
	return sv.Length
}

// et is only required for .List byte-slices.
func (sv *SliceValue) GetPointerAtIndexInt2(store Store, ii int, et Type) PointerValue {
	// Necessary run-time slice bounds check
	if ii < 0 {
		panic(fmt.Sprintf(
			"slice index out of bounds: %d", ii))
	} else if sv.Length <= ii {
		panic(fmt.Sprintf(
			"slice index out of bounds: %d (len=%d)",
			ii, sv.Length))
	}
	return sv.GetBase(store).GetPointerAtIndexInt2(store, sv.Offset+ii, et)
}

// ----------------------------------------
// StructValue

type StructValue struct {
	ObjectInfo
	Fields []TypedValue
}

// TODO handle unexported fields in debug, and also ensure in the preprocessor.
func (sv *StructValue) GetPointerTo(store Store, path ValuePath) PointerValue {
	if debug {
		if path.Depth != 0 {
			panic(fmt.Sprintf(
				"expected path.Depth of 0 but got %s %s",
				path.Name, path))
		}
	}
	return sv.GetPointerToInt(store, int(path.Index))
}

func (sv *StructValue) GetPointerToInt(store Store, index int) PointerValue {
	fv := fillValueTV(store, &sv.Fields[index])
	return PointerValue{
		TV:    fv,
		Base:  sv,
		Index: index,
	}
}

// Like GetPointerTo*, but returns (a pointer of) a reference to field.
func (sv *StructValue) GetSubrefPointerTo(store Store, st *StructType, path ValuePath) PointerValue {
	if debug {
		if path.Depth != 0 {
			panic(fmt.Sprintf(
				"expected path.Depth of 0 but got %s %s",
				path.Name, path))
		}
	}
	fv := fillValueTV(store, &sv.Fields[path.Index])
	ft := st.GetStaticTypeOfAt(path)
	return PointerValue{
		TV: &TypedValue{ // TODO: optimize
			T: &PointerType{ // TODO: optimize (cont)
				Elt: ft,
			},
			V: PointerValue{
				TV:    fv,
				Base:  sv,
				Index: int(path.Index),
			},
		},
		Base: nil, // free floating
	}
}

func (sv *StructValue) Copy(alloc *Allocator) *StructValue {
	/* TODO consider second refcount field
	if sv.GetRefCount() == 0 {
		return sv
	}
	*/
	fields := alloc.NewStructFields(len(sv.Fields))

	// Each field needs to be copied individually to ensure that
	// value fields are copied as such, even though they may be represented
	// as pointers. A good example of this would be a struct that has
	// a field that is an array. The value array is represented as a pointer.
	for i, field := range sv.Fields {
		fields[i] = field.Copy(alloc)
	}

	return alloc.NewStruct(fields)
}

// ----------------------------------------
// FuncValue

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
	Type       Type      // includes unbound receiver(s)
	IsMethod   bool      // is an (unbound) method
	Source     BlockNode // for block mem allocation
	Name       Name      // name of function/method
	Closure    Value     // *Block or RefValue to closure (may be nil for file blocks; lazy)
	FileName   Name      // file name where declared
	PkgPath    string
	NativePkg  string // for native bindings through NativeStore
	NativeName Name   // not redundant with Name; this cannot be changed in userspace

	body       []Stmt         // function body
	nativeBody func(*Machine) // alternative to Body
}

func (fv *FuncValue) IsNative() bool {
	if fv.NativePkg == "" && fv.NativeName == "" {
		return false
	}
	if fv.NativePkg == "" || fv.NativeName == "" {
		panic(fmt.Sprintf("function (%q).%s has invalid native pkg/name ((%q).%s)",
			fv.Source.GetLocation().PkgPath, fv.Name,
			fv.NativePkg, fv.NativeName))
	}
	return true
}

func (fv *FuncValue) Copy(alloc *Allocator) *FuncValue {
	alloc.AllocateFunc()
	return &FuncValue{
		Type:       fv.Type,
		IsMethod:   fv.IsMethod,
		Source:     fv.Source,
		Name:       fv.Name,
		Closure:    fv.Closure,
		FileName:   fv.FileName,
		PkgPath:    fv.PkgPath,
		NativePkg:  fv.NativePkg,
		NativeName: fv.NativeName,
		body:       fv.body,
		nativeBody: fv.nativeBody,
	}
}

func (fv *FuncValue) GetType(store Store) *FuncType {
	switch ct := fv.Type.(type) {
	case nil:
		return nil
	case RefType:
		typ := store.GetType(ct.ID).(*FuncType)
		fv.Type = typ
		return typ
	case *FuncType:
		return ct
	default:
		panic("should not happen")
	}
}

func (fv *FuncValue) GetBodyFromSource(store Store) []Stmt {
	if fv.body == nil {
		source := fv.GetSource(store)
		fv.body = source.GetBody()
		return fv.body
	}
	return fv.body
}

func (fv *FuncValue) GetSource(store Store) BlockNode {
	if rn, ok := fv.Source.(RefNode); ok {
		source := store.GetBlockNode(rn.GetLocation())
		fv.Source = source
		return source
	}
	return fv.Source
}

func (fv *FuncValue) GetPackage(store Store) *PackageValue {
	pv := store.GetPackage(fv.PkgPath, false)
	return pv
}

// NOTE: this function does not automatically memoize the closure for
// file-level declared methods and functions. For those, caller
// should set .Closure manually after *FuncValue.Copy().
func (fv *FuncValue) GetClosure(store Store) *Block {
	switch cv := fv.Closure.(type) {
	case nil:
		if fv.FileName == "" {
			return nil
		}
		pv := fv.GetPackage(store)
		fb := pv.fBlocksMap[fv.FileName]
		if fb == nil {
			panic(fmt.Sprintf("file block missing for file %q", fv.FileName))
		}
		return fb
	case RefValue:
		block := store.GetObject(cv.ObjectID).(*Block)
		fv.Closure = block
		return block
	case *Block:
		return cv
	default:
		panic("should not happen")
	}
}

// ----------------------------------------
// BoundMethodValue

type BoundMethodValue struct {
	ObjectInfo

	// Underlying unbound method function.
	// The type without the receiver (since bound)
	// is computed lazily if needed.
	Func *FuncValue

	// This becomes the first arg.
	// The type is .Func.Type.Params[0].
	Receiver TypedValue
}

// ----------------------------------------
// MapValue

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

type MapListImage struct {
	List []*MapListItem
}

func (ml MapList) MarshalAmino() (MapListImage, error) {
	mlimg := make([]*MapListItem, 0, ml.Size)
	for head := ml.Head; head != nil; head = head.Next {
		mlimg = append(mlimg, head)
	}
	return MapListImage{List: mlimg}, nil
}

func (ml *MapList) UnmarshalAmino(mlimg MapListImage) error {
	for i, item := range mlimg.List {
		if i == 0 {
			// init case
			ml.Head = item
		}
		item.Prev = ml.Tail
		if ml.Tail != nil {
			ml.Tail.Next = item
		}
		ml.Tail = item
		ml.Size++
	}
	return nil
}

// NOTE: Value is undefined until assigned.
func (ml *MapList) Append(alloc *Allocator, key TypedValue) *MapListItem {
	alloc.AllocateMapItem()
	item := &MapListItem{
		Prev: ml.Tail,
		Next: nil,
		Key:  key,
		// Value: undefined,
	}
	if ml.Head == nil {
		ml.Head = item
	}
	if ml.Tail != nil {
		ml.Tail.Next = item
	}
	ml.Tail = item
	ml.Size++
	return item
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
	Prev  *MapListItem `json:"-"`
	Next  *MapListItem `json:"-"`
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
func (mv *MapValue) GetPointerForKey(alloc *Allocator, store Store, key *TypedValue) PointerValue {
	kmk := key.ComputeMapKey(store, false)
	if mli, ok := mv.vmap[kmk]; ok {
		key2 := key.Copy(alloc)
		return PointerValue{
			TV:    fillValueTV(store, &mli.Value),
			Base:  mv,
			Key:   &key2,
			Index: PointerIndexMap,
		}
	}
	mli := mv.List.Append(alloc, *key)
	mv.vmap[kmk] = mli
	key2 := key.Copy(alloc)
	return PointerValue{
		TV:    fillValueTV(store, &mli.Value),
		Base:  mv,
		Key:   &key2,
		Index: PointerIndexMap,
	}
}

// Like GetPointerForKey, but does not create a slot if key
// doesn't exist.
func (mv *MapValue) GetValueForKey(store Store, key *TypedValue) (val TypedValue, ok bool) {
	kmk := key.ComputeMapKey(store, false)
	if mli, exists := mv.vmap[kmk]; exists {
		fillValueTV(store, &mli.Value)
		val, ok = mli.Value, true
	}
	return
}

func (mv *MapValue) DeleteForKey(store Store, key *TypedValue) {
	kmk := key.ComputeMapKey(store, false)
	if mli, ok := mv.vmap[kmk]; ok {
		mv.List.Remove(mli)
		delete(mv.vmap, kmk)
	}
}

// ----------------------------------------
// TypeValue

// The type itself as a value.
type TypeValue struct {
	Type Type
}

// ----------------------------------------
// PackageValue

type PackageValue struct {
	ObjectInfo // is a separate object from .Block.
	Block      Value
	PkgName    Name
	PkgPath    string
	FNames     []Name
	FBlocks    []Value
	Realm      *Realm `json:"-"` // if IsRealmPath(PkgPath), otherwise nil.
	// NOTE: Realm is persisted separately.

	fBlocksMap map[Name]*Block
}

func (pv *PackageValue) IsRealm() bool {
	return IsRealmPath(pv.PkgPath)
}

func (pv *PackageValue) getFBlocksMap() map[Name]*Block {
	if pv.fBlocksMap == nil {
		pv.fBlocksMap = make(map[Name]*Block, len(pv.FNames))
	}
	return pv.fBlocksMap
}

// to call after loading *PackageValue.
func (pv *PackageValue) deriveFBlocksMap(store Store) {
	if pv.fBlocksMap != nil {
		panic("should not happen")
	}
	pv.fBlocksMap = make(map[Name]*Block, len(pv.FNames))
	for i := 0; i < len(pv.FNames); i++ {
		fname := pv.FNames[i]
		fblock := pv.GetFileBlock(store, fname)
		pv.fBlocksMap[fname] = fblock
	}
}

func (pv *PackageValue) GetBlock(store Store) *Block {
	bv := pv.Block
	switch bv := bv.(type) {
	case RefValue:
		bb := store.GetObject(bv.ObjectID).(*Block)
		pv.Block = bb
		return bb
	case *Block:
		return bv
	default:
		panic("should not happen")
	}
}

func (pv *PackageValue) GetValueAt(store Store, path ValuePath) TypedValue {
	return *(pv.
		GetBlock(store).
		GetPointerTo(store, path).
		TV)
}

func (pv *PackageValue) AddFileBlock(fn Name, fb *Block) {
	for _, fname := range pv.FNames {
		if fname == fn {
			panic(fmt.Sprintf(
				"duplicate file block for file %s",
				fn))
		}
	}
	pv.FNames = append(pv.FNames, fn)
	pv.FBlocks = append(pv.FBlocks, fb)
	pv.getFBlocksMap()[fn] = fb
	fb.SetOwner(pv)
}

func (pv *PackageValue) GetFileBlock(store Store, fname Name) *Block {
	if fb, ex := pv.getFBlocksMap()[fname]; ex {
		return fb
	}
	for i, fn := range pv.FNames {
		if fn == fname {
			fbv := pv.FBlocks[i]
			switch fbv := fbv.(type) {
			case RefValue:
				fb := store.GetObject(fbv.ObjectID).(*Block)
				pv.getFBlocksMap()[fname] = fb
				return fb
			case *Block:
				pv.getFBlocksMap()[fname] = fbv
				return fbv
			default:
				panic("should not happen")
			}
		}
	}
	panic(fmt.Sprintf(
		"file %v not found in package %v",
		fname,
		pv))
}

func (pv *PackageValue) GetRealm() *Realm {
	return pv.Realm
}

func (pv *PackageValue) SetRealm(rlm *Realm) {
	pv.Realm = rlm
}

// Convenience.
func (pv *PackageValue) GetPackageNode(store Store) *PackageNode {
	return pv.GetBlock(store).GetSource(store).(*PackageNode)
}

// Convenience
func (pv *PackageValue) GetPkgAddr() crypto.Address {
	return DerivePkgAddr(pv.PkgPath)
}

// ----------------------------------------
// NativeValue

type NativeValue struct {
	Value reflect.Value `json:"-"`
	Bytes []byte        // XXX is this used?
}

func (nv *NativeValue) Copy(alloc *Allocator) *NativeValue {
	nt := nv.Value.Type()
	nv2 := reflect.New(nt).Elem()
	nv2.Set(nv.Value)
	return alloc.NewNative(nv2)
}

// ----------------------------------------
// TypedValue (is not a value, but a tuple)

type TypedValue struct {
	T Type    `json:",omitempty"` // never nil
	V Value   `json:",omitempty"` // an untyped value
	N [8]byte `json:",omitempty"` // numeric bytes
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
	}
	return tv.IsNilInterface()
}

func (tv *TypedValue) IsNilInterface() bool {
	if tv.T != nil && tv.T.Kind() == InterfaceKind {
		if tv.V == nil {
			return true
		}
		if debug {
			if tv.N != [8]byte{} {
				panic(fmt.Sprintf(
					"corrupted TypeValue (nil interface)"))
			}
		}
		return false
	}
	return false
}

func (tv *TypedValue) HasKind(k Kind) bool {
	if tv.T == nil {
		return false
	}
	return tv.T.Kind() == k
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

func (tv *TypedValue) ClearNum() {
	*(*uint64)(unsafe.Pointer(&tv.N)) = uint64(0)
}

func (tv TypedValue) Copy(alloc *Allocator) (cp TypedValue) {
	switch cv := tv.V.(type) {
	case BigintValue:
		cp.T = tv.T
		cp.V = cv.Copy(alloc)
	case *ArrayValue:
		cp.T = tv.T
		cp.V = cv.Copy(alloc)
	case *StructValue:
		cp.T = tv.T
		cp.V = cv.Copy(alloc)
	case *NativeValue:
		cp.T = tv.T
		cp.V = cv.Copy(alloc)
	default:
		cp = tv
	}
	return
}

// unrefCopy makes a copy of the underlying value in the case of reference values.
// It copies other values as expected using the normal Copy method.
func (tv TypedValue) unrefCopy(alloc *Allocator, store Store) (cp TypedValue) {
	switch tv.V.(type) {
	case RefValue:
		cp.T = tv.T
		refObject := tv.GetFirstObject(store)
		switch refObjectValue := refObject.(type) {
		case *ArrayValue:
			cp.V = refObjectValue.Copy(alloc)
		case *StructValue:
			cp.V = refObjectValue.Copy(alloc)
		default:
			cp = tv
		}
	default:
		cp = tv.Copy(alloc)
	}

	return
}

// Returns encoded bytes for primitive values.
// These bytes are used for both value hashes as well
// as hash key bytes.
func (tv *TypedValue) PrimitiveBytes() (data []byte) {
	switch bt := baseOf(tv.T); bt {
	case BoolType:
		if tv.GetBool() {
			return []byte{0x01}
		}
		return []byte{0x00}
	case StringType:
		return []byte(tv.GetString())
	case Int8Type:
		return []byte{uint8(tv.GetInt8())}
	case Int16Type:
		data = make([]byte, 2)
		binary.LittleEndian.PutUint16(
			data, uint16(tv.GetInt16()))
		return data
	case Int32Type:
		data = make([]byte, 4)
		binary.LittleEndian.PutUint32(
			data, uint32(tv.GetInt32()))
		return data
	case IntType, Int64Type:
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(
			data, uint64(tv.GetInt()))
		return data
	case Uint8Type:
		return []byte{tv.GetUint8()}
	case Uint16Type:
		data = make([]byte, 2)
		binary.LittleEndian.PutUint16(
			data, tv.GetUint16())
		return data
	case Uint32Type:
		data = make([]byte, 4)
		binary.LittleEndian.PutUint32(
			data, tv.GetUint32())
		return data
	case UintType, Uint64Type:
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(
			data, uint64(tv.GetUint()))
		return data
	case Float32Type:
		data = make([]byte, 4)
		u32 := math.Float32bits(tv.GetFloat32())
		binary.LittleEndian.PutUint32(
			data, u32)
		return data
	case Float64Type:
		data = make([]byte, 8)
		u64 := math.Float64bits(tv.GetFloat64())
		binary.LittleEndian.PutUint64(
			data, u64)
		return data
	case BigintType:
		return tv.V.(BigintValue).V.Bytes()
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
		if tv.T.Kind() != BoolKind || isNative(tv.T) {
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

func (tv *TypedValue) SetString(s StringValue) {
	if debug {
		if tv.T.Kind() != StringKind || isNative(tv.T) {
			panic(fmt.Sprintf(
				"TypedValue.SetString() on type %s",
				tv.T.String()))
		}
	}
	tv.V = s
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
	}
	return string(tv.V.(StringValue))
}

func (tv *TypedValue) SetInt(n int) {
	if debug {
		if tv.T.Kind() != IntKind || isNative(tv.T) {
			panic(fmt.Sprintf(
				"TypedValue.SetInt() on type %s",
				tv.T.String()))
		}
	}
	// XXX probably should be coerced into int64 for determinism.
	// XXX otherwise, all nodes must run in 64bit.
	// XXX alternatively, require 64bit.
	*(*int)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) ConvertGetInt() int {
	var store Store = nil // not used
	ConvertTo(nilAllocator, store, tv, IntType)
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
		if tv.T.Kind() != Int8Kind || isNative(tv.T) {
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
		if tv.T.Kind() != Int16Kind || isNative(tv.T) {
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
		if tv.T.Kind() != Int32Kind || isNative(tv.T) {
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
		if tv.T.Kind() != Int64Kind || isNative(tv.T) {
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
		if tv.T.Kind() != UintKind || isNative(tv.T) {
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
		if tv.T.Kind() != Uint8Kind || isNative(tv.T) {
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
		if tv.T != DataByteType || isNative(tv.T) {
			panic(fmt.Sprintf(
				"TypedValue.SetDataByte() on type %s",
				tv.T.String()))
		}
	}
	dbv := tv.V.(DataByteValue)
	dbv.SetByte(n)
}

func (tv *TypedValue) GetDataByte() uint8 {
	if debug {
		if tv.T != nil && tv.T != DataByteType {
			panic(fmt.Sprintf(
				"TypedValue.GetDataByte() on type %s",
				tv.T.String()))
		}
	}
	dbv := tv.V.(DataByteValue)
	return dbv.GetByte()
}

func (tv *TypedValue) SetUint16(n uint16) {
	if debug {
		if tv.T.Kind() != Uint16Kind || isNative(tv.T) {
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
		if tv.T.Kind() != Uint32Kind || isNative(tv.T) {
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
		if tv.T.Kind() != Uint64Kind || isNative(tv.T) {
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

func (tv *TypedValue) SetFloat32(n float32) {
	if debug {
		if tv.T.Kind() != Float32Kind || isNative(tv.T) {
			panic(fmt.Sprintf(
				"TypedValue.SetFloat32() on type %s",
				tv.T.String()))
		}
	}
	*(*float32)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetFloat32() float32 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Float32Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetFloat32() on type %s",
				tv.T.String()))
		}
	}
	return *(*float32)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetFloat64(n float64) {
	if debug {
		if tv.T.Kind() != Float64Kind || isNative(tv.T) {
			panic(fmt.Sprintf(
				"TypedValue.SetFloat64() on type %s",
				tv.T.String()))
		}
	}
	*(*float64)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetFloat64() float64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Float64Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetFloat64() on type %s",
				tv.T.String()))
		}
	}
	return *(*float64)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) GetBigInt() *big.Int {
	if debug {
		if tv.T != nil && tv.T.Kind() != BigintKind {
			panic(fmt.Sprintf(
				"TypedValue.GetBigInt() on type %s",
				tv.T.String()))
		}
	}
	return tv.V.(BigintValue).V
}

func (tv *TypedValue) GetBigDec() *apd.Decimal {
	if debug {
		if tv.T != nil && tv.T.Kind() != BigdecKind {
			panic(fmt.Sprintf(
				"TypedValue.GetBigDec() on type %s",
				tv.T.String()))
		}
	}
	return tv.V.(BigdecValue).V
}

func (tv *TypedValue) ComputeMapKey(store Store, omitType bool) MapKey {
	// Special case when nil: has no separator.
	if tv.T == nil {
		if debug {
			if omitType {
				panic("should not happen")
			}
		}
		return MapKey(nilStr)
	}
	// General case.
	bz := make([]byte, 0, 64)
	if !omitType {
		bz = append(bz, tv.T.TypeID().Bytes()...)
		bz = append(bz, ':') // type/value separator
	}
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		pbz := tv.PrimitiveBytes()
		bz = append(bz, pbz...)
	case *PointerType:
		ptr := uintptr(unsafe.Pointer(tv.V.(PointerValue).TV))
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
				ev := fillValueTV(store, &av.List[i])
				bz = append(bz, ev.ComputeMapKey(store, omitTypes)...)
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
			fv := fillValueTV(store, &sv.Fields[i])
			omitTypes := bt.Fields[i].Type.Kind() != InterfaceKind
			bz = append(bz, fv.ComputeMapKey(store, omitTypes)...)
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
	case *NativeType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"unexpected map key type %s",
			tv.T.String()))
	}
	return MapKey(bz)
}

// ----------------------------------------
// Value utility/manipulation functions.

// Unlike PointerValue.Assign2, does not consider DataByte or
// addressable NativeValue fields/elems.
// cu: convert untyped after assignment. pass false
// for const definitions, but true for all else.
func (tv *TypedValue) Assign(alloc *Allocator, tv2 TypedValue, cu bool) {
	if debug {
		if tv.T == DataByteType {
			// assignment to data byte types should only
			// happen via *PointerValue.Assign2().
			panic("should not happen")
		}
		if tv2.T == DataByteType {
			// tv2 will never be a DataByte, as it is
			// retrieved as value.
			panic("should not happen")
		}
	}
	*tv = tv2.Copy(alloc)
	if cu && isUntyped(tv.T) {
		ConvertUntypedTo(tv, defaultTypeOf(tv.T))
	}
}

// NOTE: Allocation for PointerValue is not immediate,
// as usually PointerValues are temporary for assignment
// or binary operations. When a pointer is to be
// allocated, *Allocator.AllocatePointer() is called separately,
// as in OpRef.
func (tv *TypedValue) GetPointerTo(alloc *Allocator, store Store, path ValuePath) PointerValue {
	if debug {
		if tv.IsUndefined() {
			panic("GetPointerTo() on undefined value")
		}
	}

	// NOTE: path will be mutated.
	// NOTE: this code segment similar to that in op_types.go
	var dtv *TypedValue
	var isPtr bool = false
	switch path.Type {
	case VPField:
		switch path.Depth {
		case 0:
			dtv = tv
		case 1:
			dtv = tv
			path.Depth = 0
		default:
			panic("should not happen")
		}
	case VPSubrefField:
		switch path.Depth {
		case 0:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
		case 1:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Depth = 0
		case 2:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Depth = 0
		case 3:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Depth = 0
		default:
			panic("should not happen")
		}
	case VPDerefField:
		switch path.Depth {
		case 0:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
		case 1:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
			path.Depth = 0
		case 2:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
			path.Depth = 0
		case 3:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
			path.Depth = 0
		default:
			panic("should not happen")
		}
	case VPDerefValMethod:
		dtv2 := tv.V.(PointerValue).TV
		dtv = &TypedValue{ // In case method is called on converted type, like ((*othertype)x).Method().
			T: tv.T.Elem(),
			V: dtv2.V,
			N: dtv2.N,
		}
		isPtr = true
		path.Type = VPValMethod
	case VPDerefPtrMethod:
		// dtv = tv.V.(PointerValue).TV
		// dtv not needed for nil receivers.
		isPtr = true
		path.Type = VPPtrMethod // XXX pseudo
	case VPDerefInterface:
		dtv = tv.V.(PointerValue).TV
		isPtr = true
		path.Type = VPInterface
	default:
		dtv = tv
	}
	if debug {
		path.Validate()
	}

	// fill dtv.V if needed.
	if dtv == nil {
		// skip, e.g. for nil pointer method receiver.
	} else {
		fillValueTV(store, dtv)
	}

	switch path.Type {
	case VPBlock:
		switch dtv.T.(type) {
		case *PackageType:
			pv := dtv.V.(*PackageValue)
			return pv.GetBlock(store).GetPointerTo(store, path)
		default:
			panic("should not happen")
		}
	case VPField:
		switch baseOf(dtv.T).(type) {
		case *StructType:
			return dtv.V.(*StructValue).GetPointerTo(store, path)
		case *TypeType:
			switch t := dtv.V.(TypeValue).Type.(type) {
			case *PointerType:
				dt := t.Elt.(*DeclaredType)
				tv := dt.GetValueAt(alloc, store, path)
				return PointerValue{
					TV:   &tv, // heap alloc
					Base: nil, // TODO: make TypeValue an object.
				}
			case *DeclaredType:
				tv := t.GetValueAt(alloc, store, path)
				return PointerValue{
					TV:   &tv, // heap alloc
					Base: nil, // TODO: make TypeValue an object.
				}
			case *NativeType:
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
				mtv := go2GnoValue(alloc, mt.Func)
				return PointerValue{
					TV:   &mtv, // heap alloc
					Base: nil,
				}
			default:
				panic("unexpected selector base typeval.")
			}
		default:
			panic(fmt.Sprintf("unexpected selector base type %s (%s)",
				dtv.T.String(), reflect.TypeOf(dtv.T)))
		}
	case VPSubrefField:
		switch ct := baseOf(dtv.T).(type) {
		case *StructType:
			return dtv.V.(*StructValue).GetSubrefPointerTo(store, ct, path)
		default:
			panic(fmt.Sprintf("unexpected (subref) selector base type %s (%s)",
				dtv.T.String(), reflect.TypeOf(dtv.T)))
		}
	case VPValMethod:
		dt := dtv.T.(*DeclaredType)
		mtv := dt.GetValueAt(alloc, store, path)
		mv := mtv.GetFunc()
		mt := mv.GetType(store)
		if debug {
			if mt.HasPointerReceiver() {
				panic("should not happen")
			}
		}
		dtv2 := dtv.Copy(alloc)
		alloc.AllocateBoundMethod()
		bmv := &BoundMethodValue{
			Func:     mv,
			Receiver: dtv2,
		}
		return PointerValue{
			TV: &TypedValue{
				T: mt.BoundType(),
				V: bmv,
			},
			Base: nil, // a bound method is free floating.
		}
	case VPPtrMethod:
		dt := tv.T.(*PointerType).Elt.(*DeclaredType)
		// ^ support nil receivers, vs:
		// dt := dtv.T.(*DeclaredType)
		mtv := dt.GetValueAt(alloc, store, path)
		mv := mtv.GetFunc()
		mt := mv.GetType(store)
		if debug {
			if !mt.HasPointerReceiver() {
				panic("should not happen")
			}
			if !isPtr {
				panic("should not happen")
			}
			if tv.T.Kind() != PointerKind {
				panic("should not happen")
			}
		}
		alloc.AllocateBoundMethod()
		bmv := &BoundMethodValue{
			Func:     mv,
			Receiver: *tv, // bound to ptr, not dtv.
		}
		return PointerValue{
			TV: &TypedValue{
				T: mt.BoundType(),
				V: bmv,
			},
			Base: nil, // a bound method is free floating.
		}
	case VPInterface:
		if dtv.IsUndefined() {
			panic("interface method call on undefined value")
		}
		callerPath := dtv.T.GetPkgPath()
		tr, _, _, _, _ := findEmbeddedFieldType(callerPath, dtv.T, path.Name, nil)
		if len(tr) == 0 {
			panic(fmt.Sprintf("method %s not found in type %s",
				path.Name, dtv.T.String()))
		}
		bv := *dtv
		for i, path := range tr {
			ptr := bv.GetPointerTo(alloc, store, path)
			if i == len(tr)-1 {
				return ptr // done
			}
			bv = ptr.Deref() // deref
		}
		panic("should not happen")
	case VPNative:
		var nv *NativeValue
		// Special case if tv.T.(PointerType):
		// we may need to treat this as a native pointer
		// to get the correct pointer-receiver value.
		if _, ok := dtv.T.(*PointerType); ok {
			pv := dtv.V.(PointerValue)
			nv = pv.TV.V.(*NativeValue)
		} else {
			nv = dtv.V.(*NativeValue)
		}
		rv := nv.Value
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
			ftv := go2GnoValue(alloc, fv)
			return PointerValue{
				TV: &ftv, // heap alloc
				// TODO consider if needed for persistence:
				/*
					Base: nv,
					Index: PointerIndexNative,
					Key: pathValue{path},
				*/
			}
		}
		// Then, try to get method.
		mv := rv.MethodByName(string(path.Name))
		if mv.IsValid() {
			mt := mv.Type()
			return PointerValue{
				TV: &TypedValue{ // heap alloc
					T: alloc.NewType(&NativeType{Type: mt}),
					V: alloc.NewNative(mv),
				},
				// TODO consider if needed for persistence:
				/*
					Base: nv,
					Index: PointerIndexNative,
					Key: pathValue{path},
				*/
			}
		} else {
			// Always try to get method from pointer type.
			if !rv.CanAddr() {
				// Replace rv with addressable value.
				rv2 := reflect.New(rt).Elem()
				rv2.Set(rv)
				rv = rv2
				tv.V.(*NativeValue).Value = rv2 // replace rv
			}
			mv := rv.Addr().MethodByName(string(path.Name))
			if mv.IsValid() {
				mt := mv.Type()
				return PointerValue{
					TV: &TypedValue{ // heap alloc
						T: alloc.NewType(&NativeType{Type: mt}),
						V: alloc.NewNative(mv),
					},
					// TODO consider if needed for persistence:
					/*
						Base: nv,
						Index: PointerIndexNative,
						Key: pathValue{path},
					*/
				}
			}
		}
		panic(fmt.Sprintf(
			"native type %s has no method or field %s",
			dtv.T.String(), path.Name))
	default:
		panic("should not happen")
	}
}

// Convenience for GetPointerAtIndex().  Slow.
func (tv *TypedValue) GetPointerAtIndexInt(store Store, ii int) PointerValue {
	iv := TypedValue{T: IntType}
	iv.SetInt(ii)
	return tv.GetPointerAtIndex(nilAllocator, store, &iv)
}

func (tv *TypedValue) GetPointerAtIndex(alloc *Allocator, store Store, iv *TypedValue) PointerValue {
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		if bt == StringType || bt == UntypedStringType {
			sv := tv.GetString()
			ii := iv.ConvertGetInt()
			bv := &TypedValue{ // heap alloc
				T: Uint8Type,
			}
			bv.SetUint8(sv[ii])
			return PointerValue{
				TV:   bv,
				Base: nil, // free floating
			}
		}
		panic(fmt.Sprintf(
			"primitive type %s cannot be indexed",
			tv.T.String()))
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		ii := iv.ConvertGetInt()
		return av.GetPointerAtIndexInt2(store, ii, bt.Elt)
	case *SliceType:
		if tv.V == nil {
			panic("nil slice index (out of bounds)")
		}
		sv := tv.V.(*SliceValue)
		ii := iv.ConvertGetInt()
		return sv.GetPointerAtIndexInt2(store, ii, bt.Elt)
	case *MapType:
		if tv.V == nil {
			panic("uninitialized map index")
		}
		mv := tv.V.(*MapValue)
		pv := mv.GetPointerForKey(alloc, store, iv)
		if pv.TV.IsUndefined() {
			vt := baseOf(tv.T).(*MapType).Value
			if vt.Kind() != InterfaceKind {
				// this will get assigned over, so no alloc.
				*(pv.TV) = defaultTypedValue(nil, vt)
			}
		}
		return pv
	case *NativeType:
		rt := tv.T.(*NativeType).Type
		nv := tv.V.(*NativeValue)
		rv := nv.Value
		switch rt.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			ii := iv.ConvertGetInt()
			erv := rv.Index(ii)
			etv := go2GnoValue(alloc, erv)
			return PointerValue{
				TV: &etv,
				// TODO consider if needed for persistence:
				/*
					Base: nv,
					Index: PointerIndexNative,
					Key: pathValue{path},
				*/
			}
		case reflect.Map:
			krv := gno2GoValue(iv, reflect.Value{})
			vrv := rv.MapIndex(krv)
			etv := go2GnoValue(alloc, vrv) // NOTE: lazy, often native.
			return PointerValue{
				TV:    &etv, // TODO not needed for assignment.
				Base:  nv,
				Index: PointerIndexNative,
				Key: &TypedValue{
					T: alloc.NewType(&NativeType{Type: krv.Type()}),
					V: alloc.NewNative(krv),
				},
			}
		default:
			panic("should not happen")
		}
	default:
		panic(fmt.Sprintf(
			"unexpected index base type %s (%v)",
			tv.T.String(),
			reflect.TypeOf(tv.T)))
	}
}

func (tv *TypedValue) SetType(tt Type) {
	tvv := tv.V.(TypeValue)
	tvv.Type = tt
	tv.V = tvv
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
	case *NativeValue:
		return cv.Value.Len()
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
	case StringValue:
		return len(string(cv))
	case *ArrayValue:
		return cv.GetCapacity()
	case *SliceValue:
		return cv.GetCapacity()
	case *NativeValue:
		return cv.Value.Cap()
	default:
		panic(fmt.Sprintf("unexpected type for cap(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetSlice(alloc *Allocator, low, high int) TypedValue {
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
		if t == StringType || t == UntypedStringType {
			return TypedValue{
				T: tv.T,
				V: alloc.NewString(tv.GetString()[low:high]),
			}
		}
		panic("non-string primitive type cannot be sliced")
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		st := alloc.NewType(&SliceType{
			Elt: t.Elt,
			Vrd: false,
		})
		return TypedValue{
			T: st,
			V: alloc.NewSlice(
				av,                   // base
				low,                  // offset
				high-low,             // length
				av.GetCapacity()-low, // maxcap
			),
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
			V: alloc.NewSlice(
				sv.Base,       // base
				sv.Offset+low, // offset
				high-low,      // length
				sv.Maxcap-low, // maxcap
			),
		}
	default:
		panic(fmt.Sprintf("unexpected type for GetSlice(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetSlice2(alloc *Allocator, low, high, max int) TypedValue {
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
		st := alloc.NewType(&SliceType{
			Elt: bt.Elt,
			Vrd: false,
		})
		return TypedValue{
			T: st,
			V: alloc.NewSlice(
				av,       // base
				low,      // low
				high-low, // length
				max-low,  // maxcap
			),
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
			V: alloc.NewSlice(
				sv.Base,       // base
				sv.Offset+low, // offset
				high-low,      // length
				max-low,       // maxcap
			),
		}
	default:
		panic(fmt.Sprintf("unexpected type for GetSlice2(): %s",
			tv.T.String()))
	}
}

// ----------------------------------------
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

// TODO rename to BlockValue.
type Block struct {
	ObjectInfo // for closures
	Source     BlockNode
	Values     []TypedValue
	Parent     Value
	Blank      TypedValue // captures "_" // XXX remove and replace with global instance.
	bodyStmt   bodyStmt   // XXX expose for persistence, not needed for MVP.
}

// NOTE: for allocation, use *Allocator.NewBlock.
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
	if len(source) > 32 {
		source = source[:32] + "..."
	}
	lines := make([]string, 0, 3)
	lines = append(lines,
		fmt.Sprintf("Block(ID:%v,Addr:%p,Source:%s,Parent:%p)",
			b.ObjectInfo.ID, b, source, b.Parent)) // XXX Parent may be RefValue{}.
	if b.Source != nil {
		if _, ok := b.Source.(RefNode); ok {
			lines = append(lines,
				fmt.Sprintf("%s(RefNode names not shown)", indent))
		} else {
			for i, n := range b.Source.GetBlockNames() {
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
	}
	return strings.Join(lines, "\n")
}

func (b *Block) GetSource(store Store) BlockNode {
	if rn, ok := b.Source.(RefNode); ok {
		source := store.GetBlockNode(rn.GetLocation())
		b.Source = source
		return source
	}
	return b.Source
}

func (b *Block) GetParent(store Store) *Block {
	switch pb := b.Parent.(type) {
	case nil:
		return nil
	case *Block:
		return pb
	case RefValue:
		block := store.GetObject(pb.ObjectID).(*Block)
		b.Parent = block
		return block
	default:
		panic("should not happen")
	}
}

func (b *Block) GetPointerToInt(store Store, index int) PointerValue {
	vv := fillValueTV(store, &b.Values[index])
	return PointerValue{
		TV:    vv,
		Base:  b,
		Index: index,
	}
}

func (b *Block) GetPointerTo(store Store, path ValuePath) PointerValue {
	if path.IsBlockBlankPath() {
		if debug {
			if path.Name != "_" {
				panic(fmt.Sprintf(
					"zero value path is reserved for \"_\", but got %s",
					path.Name))
			}
		}
		return PointerValue{
			TV:    b.GetBlankRef(),
			Base:  b,
			Index: PointerIndexBlockBlank, // -1
		}
	}
	// NOTE: For most block paths, Depth starts at 1, but
	// the generation for uverse is 0.  If path.Depth is
	// 0, it implies that b == uverse, and the condition
	// would fail as if it were 1.
	for i := uint8(1); i < path.Depth; i++ {
		b = b.GetParent(store)
	}
	return b.GetPointerToInt(store, int(path.Index))
}

// Result is used has lhs for any assignments to "_".
func (b *Block) GetBlankRef() *TypedValue {
	return &b.Blank
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams1() (pv1 PointerValue) {
	pv1 = b.GetPointerTo(nil, NewValuePathBlock(1, 0, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams2() (pv1, pv2 PointerValue) {
	pv1 = b.GetPointerTo(nil, NewValuePathBlock(1, 0, ""))
	pv2 = b.GetPointerTo(nil, NewValuePathBlock(1, 1, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams3() (pv1, pv2, pv3 PointerValue) {
	pv1 = b.GetPointerTo(nil, NewValuePathBlock(1, 0, ""))
	pv2 = b.GetPointerTo(nil, NewValuePathBlock(1, 1, ""))
	pv3 = b.GetPointerTo(nil, NewValuePathBlock(1, 2, ""))
	return
}

func (b *Block) GetBodyStmt() *bodyStmt {
	return &b.bodyStmt
}

// Used by SwitchStmt upon clause match.
func (b *Block) ExpandToSize(alloc *Allocator, size uint16) {
	if debug {
		if len(b.Values) >= int(size) {
			panic(fmt.Sprintf(
				"unexpected block size shrinkage: %v vs %v",
				len(b.Values), size))
		}
	}
	alloc.AllocateBlockItems(int64(size) - int64(len(b.Values)))
	values := make([]TypedValue, int(size))
	copy(values, b.Values)
	b.Values = values
}

// NOTE: RefValue Object methods declared in ownership.go
type RefValue struct {
	ObjectID ObjectID  `json:",omitempty"`
	Escaped  bool      `json:",omitempty"`
	PkgPath  string    `json:",omitempty"`
	Hash     ValueHash `json:",omitempty"`
}

// ----------------------------------------

func defaultStructFields(alloc *Allocator, st *StructType) []TypedValue {
	tvs := alloc.NewStructFields(len(st.Fields))
	for i, ft := range st.Fields {
		if ft.Type.Kind() != InterfaceKind {
			tvs[i].T = ft.Type
			tvs[i].V = defaultValue(alloc, ft.Type)
		}
	}
	return tvs
}

func defaultStructValue(alloc *Allocator, st *StructType) *StructValue {
	return alloc.NewStruct(
		defaultStructFields(alloc, st),
	)
}

func defaultArrayValue(alloc *Allocator, at *ArrayType) *ArrayValue {
	if at.Elt.Kind() == Uint8Kind {
		return alloc.NewDataArray(at.Len)
	}
	av := alloc.NewListArray(at.Len)
	tvs := av.List
	if et := at.Elem(); et.Kind() != InterfaceKind {
		for i := 0; i < at.Len; i++ {
			tvs[i].T = et
			tvs[i].V = defaultValue(alloc, et)
		}
	}
	return av
}

func defaultValue(alloc *Allocator, t Type) Value {
	switch ct := baseOf(t).(type) {
	case nil:
		panic("unexpected nil type")
	case *ArrayType:
		return defaultArrayValue(alloc, ct)
	case *StructType:
		return defaultStructValue(alloc, ct)
	case *SliceType:
		return nil
	case *MapType:
		return nil
	case *NativeType:
		if t.Kind() == InterfaceKind {
			return nil
		} else {
			return alloc.NewNative(
				reflect.New(ct.Type).Elem(),
			)
		}
	default:
		return nil
	}
}

func defaultTypedValue(alloc *Allocator, t Type) TypedValue {
	if t.Kind() == InterfaceKind {
		return TypedValue{}
	}
	return TypedValue{
		T: t,
		V: defaultValue(alloc, t),
	}
}

func typedInt(i int) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(i)
	return tv
}

func untypedBool(b bool) TypedValue {
	tv := TypedValue{T: UntypedBoolType}
	tv.SetBool(b)
	return tv
}

func typedRune(r rune) TypedValue {
	tv := TypedValue{T: Int32Type}
	tv.SetInt32(r)
	return tv
}

// NOTE: does not allocate; used for panics.
func typedString(s string) TypedValue {
	tv := TypedValue{T: StringType}
	tv.V = StringValue(s)
	return tv
}

func fillValueTV(store Store, tv *TypedValue) *TypedValue {
	switch cv := tv.V.(type) {
	case RefValue:
		if cv.PkgPath != "" { // load package
			tv.V = store.GetPackage(cv.PkgPath, false)
		} else { // load object
			// XXX XXX allocate object.
			tv.V = store.GetObject(cv.ObjectID)
		}
	case PointerValue:
		// As a special case, cv.Base is filled
		// and cv.TV set appropriately.
		// Alternatively, could implement
		// `PointerValue.Deref(store) *TypedValue`,
		// but for execution speed traded off for
		// loading speed, we do the following for now:
		if ref, ok := cv.Base.(RefValue); ok {
			base := store.GetObject(ref.ObjectID).(Value)
			cv.Base = base
			switch cb := base.(type) {
			case *ArrayValue:
				et := baseOf(tv.T).(*PointerType).Elt
				epv := cb.GetPointerAtIndexInt2(store, cv.Index, et)
				cv.TV = epv.TV // TODO optimize? (epv.* ignored)
			case *StructValue:
				fpv := cb.GetPointerToInt(store, cv.Index)
				cv.TV = fpv.TV // TODO optimize?
			case *BoundMethodValue:
				panic("should not happen")
			case *MapValue:
				panic("should not happen")
			case *Block:
				vpv := cb.GetPointerToInt(store, cv.Index)
				cv.TV = vpv.TV // TODO optimize?
			default:
				panic("should not happen")
			}
			tv.V = cv
		}
	default:
		// do nothing
	}
	return tv
}
