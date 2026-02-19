package gnolang

// XXX TODO address "this is wrong, for var i interface{}; &i is *interface{}."

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"unsafe"

	"github.com/cockroachdb/apd/v3"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// ----------------------------------------
// (runtime) Value

type Value interface {
	assertValue()
	String() string // for debugging

	// DeepFill returns the same value, filled.
	//
	// NOTE NOT LAZY (and potentially expensive)
	// DeepFill() is only used for synchronous recursive
	// filling before running genstd generated native bindings
	// which use Gno2GoValue().  All other filling functionality is
	// lazy, so avoid using this, and keep the logic lazy.
	//
	// NOTE must use the return value since PointerValue isn't a pointer
	// receiver, and RefValue returns another type entirely.
	DeepFill(store Store) Value

	GetShallowSize() int64
	VisitAssociated(vis Visitor) (stop bool) // for GC
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
func (*Block) assertValue()            {}
func (RefValue) assertValue()          {}
func (*HeapItemValue) assertValue()    {}

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
	_ Value = &ArrayValue{}
	_ Value = &SliceValue{}
	_ Value = &StructValue{}
	_ Value = &FuncValue{}
	_ Value = &MapValue{}
	_ Value = &BoundMethodValue{}
	_ Value = TypeValue{}
	_ Value = &PackageValue{}
	_ Value = &Block{}
	_ Value = RefValue{}
	_ Value = &HeapItemValue{}
)

// ----------------------------------------
// StringValue

type StringValue string

// ----------------------------------------
// BigintValue

type BigintValue struct {
	V *big.Int
}

func (biv BigintValue) MarshalAmino() (string, error) {
	bz, err := biv.V.MarshalText()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (biv *BigintValue) UnmarshalAmino(s string) error {
	vv := big.NewInt(0)
	err := vv.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	biv.V = vv
	return nil
}

func (biv BigintValue) Copy(alloc *Allocator) BigintValue {
	return BigintValue{V: big.NewInt(0).Set(biv.V)}
}

// ----------------------------------------
// BigdecValue

type BigdecValue struct {
	V *apd.Decimal
}

func (bdv BigdecValue) MarshalAmino() (string, error) {
	bz, err := bdv.V.MarshalText()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (bdv *BigdecValue) UnmarshalAmino(s string) error {
	vv := apd.New(0, 0)
	err := vv.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	bdv.V = vv
	return nil
}

func (bdv BigdecValue) Copy(alloc *Allocator) BigdecValue {
	cp := apd.New(0, 0)
	_, err := apd.BaseContext.Add(cp, cp, bdv.V)
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

// *(PointerValue.TypedValue) must have already become initialized, namely T
// set if a typed-nil.  Index is -1 for the shared "_" block var, and -2 for
// (gno and native) map items.
//
// A pointer constructed via a &x{} composite lit expression or constructed via
// new() or make() will have a virtual HeapItemValue as base.
//
// The Base is only nil for references to certain values that cannot
// be modified anyways, such as top level functions.
//
// Allocation for PointerValue is not immediate, as usually PointerValues are
// temporary for assignment or binary operations. When a pointer is to be
// allocated, *Allocator.AllocatePointer() is called separately, as in OpRef.
//
// Since PointerValue is used internally for assignment etc, it MUST stay
// minimal for computational efficiency.
type PointerValue struct {
	TV    *TypedValue // &Base[Index] or &Base.Index.
	Base  Value       // array/struct/block, or heapitem.
	Index int         // list/fields/values index, or -1 or -2 (see below).
}

const (
	PointerIndexBlockBlank = -1 // for the "_" identifier in blocks
	PointerIndexMap        = -2 // Base is Map, use Key.
)

func (pv *PointerValue) GetBase(store Store) Object {
	switch cbase := pv.Base.(type) {
	case nil:
		return nil
	case RefValue:
		base := store.GetObject(cbase.ObjectID)
		pv.Base = base
		return base
	case Object:
		return cbase
	default:
		panic(fmt.Sprintf("unexpected pointer base type %T", cbase))
	}
}

// cu: convert untyped; pass false for const definitions
// TODO: document as something that enables into-native assignment.
// TODO: maybe consider this as entrypoint for DataByteValue too?
func (pv PointerValue) Assign2(alloc *Allocator, store Store, rlm *Realm, tv2 TypedValue, cu bool) {
	// Special cases.
	if pv.TV.T == DataByteType {
		// Special case of DataByte into (base=*SliceValue).Data.
		pv.TV.SetDataByte(tv2.GetUint8())
		return
	}
	// General case
	if rlm != nil {
		if debug && pv.Base == nil {
			panic("expected non-nil base for assignment")
		}
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
	btv := &TypedValue{ // heap alloc, so need to compare value rather than pointer
		T: DataByteType,
		V: DataByteValue{
			Base:     av,
			Index:    ii,
			ElemType: et,
		},
	}

	return PointerValue{
		TV:    btv,
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
		excpt := &Exception{
			Value: typedString(fmt.Sprintf(
				"slice index out of bounds: %d", ii)),
		}
		panic(excpt)
	} else if sv.Length <= ii {
		excpt := &Exception{
			Value: typedString(fmt.Sprintf(
				"slice index out of bounds: %d (len=%d)",
				ii, sv.Length)),
		}
		panic(excpt)
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
	ObjectInfo
	Type       Type         // includes unbound receiver(s)
	IsMethod   bool         // is an (unbound) method
	IsClosure  bool         // is a func lit expr closure (not decl)
	Source     BlockNode    // for block mem allocation
	Name       Name         // name of function/method
	Parent     Value        // *Block or RefValue to closure (may be nil for file blocks; lazy)
	Captures   []TypedValue `json:",omitempty"` // HeapItemValues captured from closure.
	FileName   string       // file name where declared
	PkgPath    string       // package path in which func declared
	NativePkg  string       // for native bindings through NativeResolver
	NativeName Name         // not redundant with Name; this cannot be changed in userspace
	Crossing   bool         // true if .body's first statement is crossing().

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
		Parent:     fv.Parent,
		FileName:   fv.FileName,
		PkgPath:    fv.PkgPath,
		NativePkg:  fv.NativePkg,
		NativeName: fv.NativeName,
		Crossing:   fv.Crossing,
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

func (fv *FuncValue) UpdateBodyFromSource() {
	if fv.Source == nil {
		panic(fmt.Sprintf(
			"Source is missing for FuncValue %q",
			fv.Name))
	}
	fv.body = fv.Source.GetBody()
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

func (fv *FuncValue) GetParent(store Store) *Block {
	if fv.IsClosure {
		return nil
	}
	switch cv := fv.Parent.(type) {
	case nil:
		if fv.FileName == "" {
			return nil
		}
		pv := fv.GetPackage(store)
		fb, ok := pv.fBlocksMap[fv.FileName]
		if !ok {
			panic(fmt.Sprintf("file block missing for file %q", fv.FileName))
		}
		fv.Parent = fb
		return fb
	case RefValue:
		block := store.GetObject(cv.ObjectID).(*Block)
		fv.Parent = block
		return block
	case *Block:
		return cv
	default:
		panic("should not happen")
	}
}

func (fv *FuncValue) IsCrossing() bool {
	return fv.Crossing
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

func (bmv *BoundMethodValue) IsCrossing() bool {
	return bmv.Func.IsCrossing()
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
			ml.Head = item
			ml.Tail = item
			item.Prev = nil
		} else {
			item.Prev = ml.Tail
			ml.Tail.Next = item
			ml.Tail = item
		}
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
		ml.Tail = item
	} else {
		ml.Tail.Next = item
		ml.Tail = item
	}
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

// GetPointerForKey is only used for assignment, so the key
// is not returned as part of the pointer, and TV is not filled.
func (mv *MapValue) GetPointerForKey(alloc *Allocator, store Store, key TypedValue) PointerValue {
	// If NaN, instead of computing map key, just append to List.
	kmk, isNaN := key.ComputeMapKey(store, false)
	if !isNaN {
		if mli, ok := mv.vmap[kmk]; ok {
			// When assigning to a map item, the key is always equal to that of the
			// last assignment; this is mostly noticeable in the case of -0 / 0:
			// https://go.dev/play/p/iNPDR4FQlRv
			mli.Key = key
			return PointerValue{
				TV:    &mli.Value,
				Base:  mv,
				Index: PointerIndexMap,
			}
		}
	}
	mli := mv.List.Append(alloc, key)

	mv.vmap[kmk] = mli
	return PointerValue{
		TV:    &mli.Value,
		Base:  mv,
		Index: PointerIndexMap,
	}
}

// Like GetPointerForKey, but does not create a slot if key
// doesn't exist.
func (mv *MapValue) GetValueForKey(store Store, key *TypedValue) (val TypedValue, ok bool) {
	// If key is NaN, return default
	kmk, isNaN := key.ComputeMapKey(store, false)
	if isNaN {
		return
	}
	if mli, exists := mv.vmap[kmk]; exists {
		fillValueTV(store, &mli.Value)
		val, ok = mli.Value, true
	}
	return
}

func (mv *MapValue) DeleteForKey(store Store, key *TypedValue) {
	// if key is NaN, do nothing.
	kmk, isNaN := key.ComputeMapKey(store, false)
	if isNaN {
		return
	}
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
	FNames     []string
	FBlocks    []Value
	Realm      *Realm `json:"-"` // if IsRealmPath(PkgPath), otherwise nil.
	Private    bool
	// NOTE: Realm is persisted separately.

	fBlocksMap map[string]*Block
}

// See PackageNode.NewPackage() for PackageValue constructor.
// Do not uncomment; always use PackageNode.NewPackage().
// func NewPackageValue() {}

// IsRealm returns true if pv represents a realm.
func (pv *PackageValue) IsRealm() bool {
	return IsRealmPath(pv.PkgPath)
}

// XXX, pass in allocator
func (pv *PackageValue) getFBlocksMap() map[string]*Block {
	if pv.fBlocksMap == nil {
		pv.fBlocksMap = make(map[string]*Block, len(pv.FNames))
	}
	return pv.fBlocksMap
}

// called after loading *PackageValue.
func (pv *PackageValue) deriveFBlocksMap(store Store) {
	if pv.fBlocksMap == nil {
		pv.fBlocksMap = make(map[string]*Block, len(pv.FNames))
	}
	for i := range pv.FNames {
		fname := pv.FNames[i]
		fblock := pv.GetFileBlock(store, fname)
		pv.fBlocksMap[fname] = fblock // idempotent.
	}
}

// Retrieves the block from store if necessary, and if so fills all the values
// of the block.
func (pv *PackageValue) GetBlock(store Store) *Block {
	bv := pv.Block
	switch bv := bv.(type) {
	case RefValue:
		bb := store.GetObject(bv.ObjectID).(*Block)
		pv.Block = bb
		for i := range bb.Values {
			tv := &bb.Values[i]
			fillValueTV(store, tv)
		}
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

func (pv *PackageValue) AddFileBlock(fname string, fb *Block) {
	for _, fn := range pv.FNames {
		if fname == fn {
			panic(fmt.Sprintf(
				"duplicate file block for file %s",
				fname))
		}
	}
	pv.FNames = append(pv.FNames, fname)
	pv.FBlocks = append(pv.FBlocks, fb)
	pv.getFBlocksMap()[fname] = fb
	fb.SetOwner(pv)
}

func (pv *PackageValue) GetFileBlock(store Store, fname string) *Block {
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

func (pv *PackageValue) SetPrivate(private bool) {
	pv.Private = private
}

// Convenience.
func (pv *PackageValue) GetPackageNode(store Store) *PackageNode {
	return pv.GetBlock(store).GetSource(store).(*PackageNode)
}

// Convenience
func (pv *PackageValue) GetPkgAddr() crypto.Address {
	return DerivePkgCryptoAddr(pv.PkgPath)
}

func (pv *PackageValue) GetPkgStorageDepositAddr() crypto.Address {
	return DeriveStorageDepositCryptoAddr(pv.PkgPath)
}

// ----------------------------------------
// TypedValue (is not a value, but a tuple)

type TypedValue struct {
	T Type    `json:",omitempty"`
	V Value   `json:",omitempty"`
	N [8]byte `json:",omitempty"`
}

// Magic 8 bytes to denote a readonly wrapped non-nil V of mutable type that is
// readonly. This happens when subvalues are retrieved from an externally
// stored realm value, such as external realm package vars, or slices or
// pointers to.
// NOTE: most of the code except copy methods do not consider N_Readonly.
// Instead the op functions should with m.IsReadonly() and tv.SetReadonly() and
// tv.WithReadonly().
var N_Readonly [8]byte = [8]byte{'R', 'e', 'a', 'D', 'o', 'N', 'L', 'Y'} // ReaDoNLY

// Returns true if mutable .V is readonly "wrapped".
func (tv *TypedValue) IsReadonly() bool {
	return tv.N == N_Readonly && tv.V != nil
}

// Sets tv.N to N_Readonly if ro and tv is not already immutable.  If ro is
// false does nothing. See also Type.IsImmutable().
func (tv *TypedValue) SetReadonly(ro bool) {
	if tv.V == nil {
		return // do nothing
	}
	if tv.T.IsImmutable() {
		return // do nothing
	}
	if ro {
		tv.N = N_Readonly
		return
	} else {
		return // preserve prior tv.N
	}
}

// Convenience, makes readonly if ro is true.
func (tv TypedValue) WithReadonly(ro bool) TypedValue {
	tv.SetReadonly(ro)
	return tv
}

func (tv *TypedValue) IsImmutable() bool {
	return tv.T == nil || tv.T.IsImmutable()
}

func (tv *TypedValue) IsDefined() bool {
	return !tv.IsUndefined()
}

func (tv *TypedValue) IsUndefined() bool {
	if debug {
		if tv.T == nil {
			if tv.V != nil {
				panic("should not happen")
			}
			if tv.N != [8]byte{} {
				panic("should not happen")
			}
		}
	}
	return tv.T == nil
}

func (tv *TypedValue) IsTypedNil() bool {
	if tv.V != nil {
		return false
	}
	if tv.T != nil {
		switch tv.T.Kind() {
		case SliceKind, FuncKind, MapKind, InterfaceKind, PointerKind, ChanKind:
			return true
		}
	}
	return false
}

// (this is used mostly by the preprocessor)
func (tv *TypedValue) IsNilInterface() bool {
	if tv.T != nil && tv.T.Kind() == InterfaceKind {
		if tv.V == nil {
			return true
		}
		if debug {
			if tv.N != [8]byte{} {
				panic("corrupted TypeValue (nil interface)")
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
		cp.N = tv.N // preserve N_Readonly
	case *StructValue:
		cp.T = tv.T
		cp.V = cv.Copy(alloc)
		cp.N = tv.N // preserve N_Readonly
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
		cp = tv // preserve N_Readonly
		refObject := tv.GetFirstObject(store)
		switch refObjectValue := refObject.(type) {
		case *ArrayValue:
			cp.V = refObjectValue.Copy(alloc)
		case *StructValue:
			cp.V = refObjectValue.Copy(alloc)
		}
	default:
		cp = tv.Copy(alloc)
	}

	return
}

// Returns encoded bytes for primitive values.
// These are used for computing map keys.
// If NaN float32|float64 returns bz unchanged and isNaN=true.
func (tv *TypedValue) MapKeyBytes(bz []byte) (bz2 []byte, isNaN bool) {
	switch bt := baseOf(tv.T); bt {
	case BoolType:
		if tv.GetBool() {
			return append(bz, 0x01), false
		}
		return append(bz, 0x00), false
	case StringType:
		return append(bz, tv.GetString()...), false
	case Int8Type:
		return append(bz, uint8(tv.GetInt8())), false
	case Int16Type:
		return binary.LittleEndian.AppendUint16(bz, uint16(tv.GetInt16())), false
	case Int32Type:
		return binary.LittleEndian.AppendUint32(bz, uint32(tv.GetInt32())), false
	case IntType, Int64Type:
		return binary.LittleEndian.AppendUint64(bz, uint64(tv.GetInt())), false
	case Uint8Type:
		return append(bz, tv.GetUint8()), false
	case Uint16Type:
		return binary.LittleEndian.AppendUint16(bz, tv.GetUint16()), false
	case Uint32Type:
		return binary.LittleEndian.AppendUint32(bz, tv.GetUint32()), false
	case UintType, Uint64Type:
		return binary.LittleEndian.AppendUint64(bz, tv.GetUint()), false
	case Float32Type:
		u32 := tv.GetFloat32()
		if u32 == 0 || u32 == (1<<31) { // 0 or -0 normalized to 0
			return binary.LittleEndian.AppendUint32(bz, 0), false
		}
		if _, _, _, _, isNaN = softfloat.Funpack32(u32); isNaN {
			return bz, true
		}
		return binary.LittleEndian.AppendUint32(bz, u32), false
	case Float64Type:
		u64 := tv.GetFloat64()
		if u64 == 0 || u64 == (1<<63) { // 0 or -0 normalized to 0
			return binary.LittleEndian.AppendUint64(bz, 0), false
		}
		if _, _, _, _, isNaN = softfloat.Funpack64(u64); isNaN {
			return bz, true
		}
		return binary.LittleEndian.AppendUint64(bz, u64), false
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

func (tv *TypedValue) SetString(s StringValue) {
	if debug {
		if tv.T.Kind() != StringKind {
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

func (tv *TypedValue) SetInt(n int64) {
	if debug {
		if tv.T.Kind() != IntKind {
			panic(fmt.Sprintf(
				"TypedValue.SetInt() on type %s",
				tv.T.String()))
		}
	}
	*(*int64)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) ConvertGetInt() int64 {
	var store Store = nil // not used
	ConvertTo(nilAllocator, store, tv, IntType, false)
	return tv.GetInt()
}

func (tv *TypedValue) GetInt() int64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != IntKind {
			panic(fmt.Sprintf(
				"TypedValue.GetInt() on type %s",
				tv.T.String()))
		}
	}
	return *(*int64)(unsafe.Pointer(&tv.N))
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

func (tv *TypedValue) SetUint(n uint64) {
	if debug {
		if tv.T.Kind() != UintKind {
			panic(fmt.Sprintf(
				"TypedValue.SetUint() on type %s",
				tv.T.String()))
		}
	}
	*(*uint64)(unsafe.Pointer(&tv.N)) = n
}

func (tv *TypedValue) GetUint() uint64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != UintKind {
			panic(fmt.Sprintf(
				"TypedValue.GetUint() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint64)(unsafe.Pointer(&tv.N))
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

func (tv *TypedValue) SetFloat32(u32 uint32) {
	if debug {
		if tv.T.Kind() != Float32Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetFloat32() on type %s",
				tv.T.String()))
		}
	}
	*(*uint32)(unsafe.Pointer(&tv.N)) = u32
}

func (tv *TypedValue) GetFloat32() uint32 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Float32Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetFloat32() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint32)(unsafe.Pointer(&tv.N))
}

func (tv *TypedValue) SetFloat64(u64 uint64) {
	if debug {
		if tv.T.Kind() != Float64Kind {
			panic(fmt.Sprintf(
				"TypedValue.SetFloat64() on type %s",
				tv.T.String()))
		}
	}
	*(*uint64)(unsafe.Pointer(&tv.N)) = u64
}

func (tv *TypedValue) GetFloat64() uint64 {
	if debug {
		if tv.T != nil && tv.T.Kind() != Float64Kind {
			panic(fmt.Sprintf(
				"TypedValue.GetFloat64() on type %s",
				tv.T.String()))
		}
	}
	return *(*uint64)(unsafe.Pointer(&tv.N))
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

// Sign returns the sign of the given numeric tv.
// Floating points are not correctly handled for NaN values.
func (tv *TypedValue) Sign() int {
	if tv.T == nil {
		panic("type should not be nil")
	}

	switch tv.T.Kind() {
	case IntKind:
		return signOfInteger(tv.GetInt())
	case Int8Kind:
		return signOfInteger(int64(tv.GetInt8()))
	case Int16Kind:
		return signOfInteger(int64(tv.GetInt16()))
	case Int32Kind:
		return signOfInteger(int64(tv.GetInt32()))
	case Int64Kind:
		return signOfInteger(tv.GetInt64())
	case UintKind:
		return signOfInteger(tv.GetUint())
	case Uint8Kind:
		return signOfInteger(uint64(tv.GetUint8()))
	case Uint16Kind:
		return signOfInteger(uint64(tv.GetUint16()))
	case Uint32Kind:
		return signOfInteger(uint64(tv.GetUint32()))
	case Uint64Kind:
		return signOfInteger(tv.GetUint64())
	case Float32Kind:
		return signOfFloat32Bits(tv.GetFloat32())
	case Float64Kind:
		return signOfFloat64Bits(tv.GetFloat64())
	case BigintKind:
		v := tv.GetBigInt()
		return v.Sign()
	case BigdecKind:
		v := tv.GetBigDec()
		return v.Sign()
	default:
		panic("type should be numeric")
	}
}

func (tv *TypedValue) AssertNonNegative(msg string) {
	if tv.Sign() < 0 {
		panic(fmt.Sprintf("%s: %v", msg, tv))
	}
}

// ComputeMapKey returns the value of tv, encoded as a string for usage inside
// of a map.
//
// isNaN returns whether tv, or any of the values contained within (like in an
// array or struct) are NaN's; this would make the same tv != to itself, and
// so shouldn't be included within a vmap.
func (tv *TypedValue) ComputeMapKey(store Store, omitType bool) (key MapKey, isNaN bool) {
	// Special case when nil: has no separator.
	if tv.T == nil {
		if debug {
			if omitType {
				panic("should not happen")
			}
		}
		return nilStr, false
	}
	// General case.
	bz := make([]byte, 0, 64)
	if !omitType {
		// TypeID is human readable and balanced, so appending ":" works.
		// This keeps ComputeMapKey somewhat human readable esp w/
		// colors.ColoredBytes().
		bz = append(bz, tv.T.TypeID().Bytes()...)
		bz = append(bz, ':') // type/value separator
	}
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		bz, isNaN = tv.MapKeyBytes(bz)
		if isNaN {
			return "", true
		}
	case *PointerType:
		var ptrBytes [sizeOfUintPtr]byte // zero-initialized for nil pointers
		if tv.V != nil {
			ptr := uintptr(unsafe.Pointer(tv.V.(PointerValue).TV))
			ptrBytes = uintptrToBytes(&ptr)
		}
		bz = append(bz, ptrBytes[:]...)
	case FieldType:
		panic("field (pseudo)type cannot be used as map key")
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		al := av.GetLength()
		bz = append(bz, '[')
		if av.Data == nil {
			if tv.T.Elem().Kind() == Uint8Kind {
				panic("should not happen; unexpected list array for array with uint8 element kind")
			}
			omitTypes := bt.Elem().Kind() != InterfaceKind
			for i := range al {
				ev := fillValueTV(store, &av.List[i])
				mk, isNaN := ev.ComputeMapKey(store, omitTypes)
				if isNaN {
					return "", true
				}
				bz = binary.AppendUvarint(bz, uint64(len(mk)))
				bz = append(bz, mk...)
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
		for i := range sl {
			fv := fillValueTV(store, &sv.Fields[i])
			omitTypes := bt.Fields[i].Type.Kind() != InterfaceKind
			mk, isNaN := fv.ComputeMapKey(store, omitTypes)
			if isNaN {
				return "", true
			}
			bz = binary.AppendUvarint(bz, uint64(len(mk)))
			bz = append(bz, mk...)
			if i != sl-1 {
				bz = append(bz, ',')
			}
		}
		bz = append(bz, '}')
	case *ChanType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"unexpected map key type %s",
			tv.T.String()))
	}
	return MapKey(bz), false
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

// Define to a block slot that takes into account heap escapes.
// (only blocks can contain heap items).
// This should only be used when both the base parent and the value are unreal
// new values, or call rlm.DidUpdate manually.
func (tv *TypedValue) AssignToBlock(other TypedValue) {
	if _, ok := tv.T.(heapItemType); ok {
		tv.V.(*HeapItemValue).Value = other
	} else {
		*tv = other
	}
}

// Like AssignToBlock but creates a new heap item instead.
// This should only be used when both the base parent and the value are unreal
// new values, or call rlm.DidUpdate manually.
func (tv *TypedValue) DefineToBlock(other TypedValue) {
	if _, ok := tv.T.(heapItemType); ok {
		*tv = TypedValue{
			T: heapItemType{},
			V: &HeapItemValue{Value: other},
		}
	} else {
		*tv = other
	}
}

// NOTE: Allocation for PointerValue is not immediate,
// as usually PointerValues are temporary for assignment
// or binary operations. When a pointer is to be
// allocated, *Allocator.AllocatePointer() is called separately,
// as in OpRef.
func (tv *TypedValue) GetPointerToFromTV(alloc *Allocator, store Store, path ValuePath) PointerValue {
	if debug {
		if tv.IsUndefined() {
			panic("GetPointerToFromTV() on undefined value")
		}
	}

	// NOTE: path will be mutated.
	// NOTE: this code segment similar to that in op_types.go
	var dtv *TypedValue
	isPtr := false
	switch path.Type {
	case VPField:
		switch path.Depth {
		case 0:
			dtv = tv
		case 1:
			dtv = tv
			path.SetDepth(0)
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
			path.SetDepth(0)
		case 2:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.SetDepth(0)
		case 3:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.SetDepth(0)
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
			path.SetDepth(0)
		case 2:
			if tv.V == nil {
				panic(&Exception{Value: typedString("nil pointer dereference")})
			}
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
			path.SetDepth(0)
		case 3:
			dtv = tv.V.(PointerValue).TV
			isPtr = true
			path.Type = VPField
			path.SetDepth(0)
		default:
			panic("should not happen")
		}
	case VPDerefValMethod:
		if tv.V == nil {
			panic(&Exception{Value: typedString("nil pointer dereference")})
		}
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
		// dtv not used due to possible nil receivers.
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
		if dtv2.V != nil {
			// Clear readonly for receivers.
			// Other rules still apply such as in DidUpdate.
			// NOTE: dtv2 is a copy, orig is untouched.
			dtv2.N = [8]byte{}
		}
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
		ptv := *tv
		if ptv.V != nil {
			// Clear readonly for receivers.
			// Other rules still apply such as in DidUpdate.
			// NOTE: ptv is a copy, orig is untouched.
			ptv.N = [8]byte{}
		}
		alloc.AllocateBoundMethod()
		bmv := &BoundMethodValue{
			Func:     mv,
			Receiver: ptv, // bound to tv ptr, not dtv.
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
		if dtv.T.Kind() == InterfaceKind {
			panic("cannot resolve an interface path at static time")
		}
		callerPath := dtv.T.GetPkgPath()
		tr, _, _, _, _ := findEmbeddedFieldType(callerPath, dtv.T, path.Name, nil)
		if len(tr) == 0 {
			panic(fmt.Sprintf("method %s not found in type %s",
				path.Name, dtv.T.String()))
		}
		btv := *dtv
		for i, path := range tr {
			ptr := btv.GetPointerToFromTV(alloc, store, path)
			if i == len(tr)-1 {
				return ptr // done
			}
			btv = ptr.Deref() // deref
		}
		panic("should not happen")
	default:
		panic("should not happen")
	}
}

// Convenience for GetPointerAtIndex(). Slow.
func (tv *TypedValue) GetPointerAtIndexInt(store Store, ii int) PointerValue {
	iv := TypedValue{T: IntType}
	iv.SetInt(int64(ii))
	return tv.GetPointerAtIndex(nilRealm, nilAllocator, store, &iv)
}

func (tv *TypedValue) GetPointerAtIndex(rlm *Realm, alloc *Allocator, store Store, iv *TypedValue) PointerValue {
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		if bt == StringType || bt == UntypedStringType {
			sv := tv.GetString()
			ii := int(iv.ConvertGetInt())
			btv := &TypedValue{ // heap alloc
				T: Uint8Type,
			}

			if ii >= len(sv) {
				panic(&Exception{Value: typedString(fmt.Sprintf("index out of range [%d] with length %d", ii, len(sv)))})
			}
			if ii < 0 {
				panic(&Exception{Value: typedString(fmt.Sprintf("invalid slice index %d (index must be non-negative)", ii))})
			}

			btv.SetUint8(sv[ii])
			return PointerValue{
				TV:   btv,
				Base: nil, // free floating
			}
		}
		panic(fmt.Sprintf(
			"primitive type %s cannot be indexed",
			tv.T.String()))
	case *ArrayType:
		av := tv.V.(*ArrayValue)
		ii := int(iv.ConvertGetInt())
		return av.GetPointerAtIndexInt2(store, ii, bt.Elt)
	case *SliceType:
		if tv.V == nil {
			panic("nil slice index (out of bounds)")
		}
		sv := tv.V.(*SliceValue)
		ii := int(iv.ConvertGetInt())
		return sv.GetPointerAtIndexInt2(store, ii, bt.Elt)
	case *MapType:
		if tv.V == nil {
			panic(&Exception{Value: typedString("uninitialized map index")})
		}
		mv := tv.V.(*MapValue)

		// if a key is getting attached, we should update it with the new one,
		// as that is the one that matters. this is mostly relevant for -0 / 0.
		// https://github.com/gnolang/gno/pull/4114
		var oldObject Object
		key, isNaN := iv.ComputeMapKey(store, false)
		if !isNaN {
			k, ok := mv.vmap[key]
			if ok {
				oldObject = k.Key.GetFirstObject(store)
			}
		}

		ivk := iv.Copy(alloc)
		pv := mv.GetPointerForKey(alloc, store, ivk)
		if pv.TV.IsUndefined() {
			vt := baseOf(tv.T).(*MapType).Value
			if vt.Kind() != InterfaceKind {
				// this will get assigned over, so no alloc.
				*(pv.TV) = defaultTypedValue(nil, vt)
			}
		}
		// attach mapkey object, if changed
		newObject := ivk.GetFirstObject(store)
		if oldObject != newObject {
			rlm.DidUpdate(mv, oldObject, newObject)
		}

		return pv
	default:
		panic(fmt.Sprintf(
			"unexpected index base type %s (%v base %v)",
			tv.T.String(),
			reflect.TypeOf(tv.T),
			reflect.TypeOf(baseOf(tv.T))))
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

func (tv *TypedValue) GetUnboundFunc() *FuncValue {
	switch fv := tv.V.(type) {
	case *FuncValue:
		return fv
	case *BoundMethodValue:
		return fv.Func
	default:
		panic(fmt.Sprintf("expected function or bound method but got %T", tv.V))
	}
}

func (tv *TypedValue) GetLength() int {
	if tv.V == nil {
		switch bt := baseOf(tv.T).(type) {
		case PrimitiveType:
			if bt != StringType {
				panic(fmt.Sprintf("unexpected type for len(): %s", tv.T.String()))
			}
			return 0
		case *ArrayType:
			return bt.Len
		case *SliceType:
			return 0
		case *MapType:
			return 0
		case *PointerType:
			if at, ok := bt.Elt.(*ArrayType); ok {
				return at.Len
			}
			panic(fmt.Sprintf("unexpected type for len(): %s", tv.T.String()))
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
	case PointerValue:
		if av, ok := cv.TV.V.(*ArrayValue); ok {
			return av.GetLength()
		}
		panic(fmt.Sprintf("unexpected type for len(): %s", tv.T.String()))
	default:
		panic(fmt.Sprintf("unexpected type for len(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetCapacity() int {
	if tv.V == nil {
		// assert acceptable type.
		switch bt := baseOf(tv.T).(type) {
		// strings have no capacity.
		case *ArrayType:
			return bt.Len
		case *SliceType:
			return 0
		case *PointerType:
			if at, ok := bt.Elt.(*ArrayType); ok {
				return at.Len
			}
			panic(fmt.Sprintf("unexpected type for cap(): %s", tv.T.String()))
		default:
			panic(fmt.Sprintf("unexpected type for cap(): %s", tv.T.String()))
		}
	}
	switch cv := tv.V.(type) {
	case *ArrayValue:
		return cv.GetCapacity()
	case *SliceValue:
		return cv.GetCapacity()
	case PointerValue:
		if av, ok := cv.TV.V.(*ArrayValue); ok {
			return av.GetCapacity()
		}
		panic(fmt.Sprintf("unexpected type for cap(): %s", tv.T.String()))
	default:
		panic(fmt.Sprintf("unexpected type for cap(): %s",
			tv.T.String()))
	}
}

func (tv *TypedValue) GetSlice(alloc *Allocator, low, high int) TypedValue {
	if low < 0 {
		panic(&Exception{Value: typedString(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			low))})
	}
	if high < 0 {
		panic(&Exception{Value: typedString(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			low))})
	}
	if low > high {
		panic(&Exception{Value: typedString(fmt.Sprintf(
			"invalid slice index %d > %d",
			low, high))})
	}
	switch t := baseOf(tv.T).(type) {
	case PrimitiveType:
		if tv.GetLength() < high {
			panic(&Exception{Value: typedString(fmt.Sprintf(
				"slice bounds out of range [%d:%d] with string length %d",
				low, high, tv.GetLength()))})
		}
		if t == StringType || t == UntypedStringType {
			return TypedValue{
				T: tv.T,
				V: alloc.NewString(tv.GetString()[low:high]),
			}
		}
		panic(&Exception{Value: typedString(
			"non-string primitive type cannot be sliced",
		)})
	case *ArrayType:
		if tv.GetLength() < high {
			panic(&Exception{Value: typedString(fmt.Sprintf(
				"slice bounds out of range [%d:%d] with array length %d",
				low, high, tv.GetLength()))})
		}
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
		// XXX consider restricting slice expansion if slice is readonly.
		if tv.GetCapacity() < high {
			panic(&Exception{Value: typedString(fmt.Sprintf(
				"slice bounds out of range [%d:%d] with capacity %d",
				low, high, tv.GetCapacity()))})
		}
		if tv.V == nil {
			if low != 0 || high != 0 {
				panic(&Exception{Value: typedString("nil slice index out of range")})
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

func (tv *TypedValue) GetSlice2(alloc *Allocator, lowVal, highVal, maxVal int) TypedValue {
	if lowVal < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			lowVal))
	}
	if highVal < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			highVal))
	}
	if maxVal < 0 {
		panic(fmt.Sprintf(
			"invalid slice index %d (index must be non-negative)",
			maxVal))
	}
	if lowVal > highVal {
		panic(fmt.Sprintf(
			"invalid slice index %d > %d",
			lowVal, highVal))
	}
	if highVal > maxVal {
		panic(fmt.Sprintf(
			"invalid slice index %d > %d",
			highVal, maxVal))
	}
	if tv.GetCapacity() < highVal {
		panic(fmt.Sprintf(
			"slice bounds out of range [%d:%d:%d] with capacity %d",
			lowVal, highVal, maxVal, tv.GetCapacity()))
	}
	if tv.GetCapacity() < maxVal {
		panic(fmt.Sprintf(
			"slice bounds out of range [%d:%d:%d] with capacity %d",
			lowVal, highVal, maxVal, tv.GetCapacity()))
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
				av,             // base
				lowVal,         // low
				highVal-lowVal, // length
				maxVal-lowVal,  // maxcap
			),
		}
	case *SliceType:
		// XXX consider restricting slice expansion if slice is readonly.
		if tv.V == nil {
			if lowVal != 0 || highVal != 0 || maxVal != 0 {
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
				sv.Base,          // base
				sv.Offset+lowVal, // offset
				highVal-lowVal,   // length
				maxVal-lowVal,    // maxcap
			),
		}
	default:
		panic(fmt.Sprintf("unexpected type for GetSlice2(): %s",
			tv.T.String()))
	}
}

// Convenience for Value.DeepFill.
// NOTE: NOT LAZY (and potentially expensive)
func (tv *TypedValue) DeepFill(store Store) {
	if tv.V != nil {
		tv.V = tv.V.DeepFill(store)
	}
}

// ----------------------------------------
// Block
//
// Blocks hold values referred to by var/const/func/type declarations in
// BlockNodes such as packages, functions, and switch statements.  Unlike
// structs or packages, names and paths may refer to parent blocks.  (In the
// future, the same mechanism may be used to support inheritance or
// prototype-like functionality for structs and packages.)
//
// When a block would otherwise become gc'd because it is no longer used, the
// block is forgotten and GC'd.
//
// Variables declared in a closure or passed by reference are first discovered
// and marked as such from the preprocessor, and NewBlock() will prepopulate
// these slots with *HeapItemValues.  When a *HeapItemValue (or sometimes
// HeapItemType in .T) is present in a block slot it is not written over but
// instead the value is written into the heap item's slot--except for loopvars
// assignments which may replace the heap item with another one. This is
// how Gno supports Go1.22 loopvars.
// TODO XXX rename to BlockValue
type Block struct {
	ObjectInfo
	Source   BlockNode
	Values   []TypedValue
	Parent   Value
	Blank    TypedValue // captures "_" // XXX remove and replace with global instance.
	bodyStmt bodyStmt   // XXX expose for persistence, not needed for MVP.
}

// NOTE: for allocation, use *Allocator.NewBlock.
func NewBlock(alloc *Allocator, source BlockNode, parent *Block) *Block {
	numNames := source.GetNumNames()
	values := make([]TypedValue, numNames)
	// Keep in sync with ExpandWith().
	for i, isHeap := range source.GetHeapItems() {
		if !isHeap {
			continue
		}
		// Indicates must always be heap item.
		values[i] = TypedValue{
			T: heapItemType{},
			V: alloc.NewHeapItem(TypedValue{}),
		}
	}
	return &Block{
		Source: source,
		Values: values,
		Parent: parent,
	}
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
	if hiv, ok := vv.V.(*HeapItemValue); ok {
		fillValueTV(store, &hiv.Value)
		return PointerValue{
			TV:    &hiv.Value,
			Base:  vv.V,
			Index: 0,
		}
	} else {
		return PointerValue{
			TV:    vv,
			Base:  b,
			Index: index,
		}
	}
}

func (b *Block) GetPointerToIntDirect(store Store, index int) PointerValue {
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
			if path.Name != blankIdentifier {
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

func (b *Block) GetPointerToDirect(store Store, path ValuePath) PointerValue {
	if path.IsBlockBlankPath() {
		if debug {
			if path.Name != blankIdentifier {
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
	return b.GetPointerToIntDirect(store, int(path.Index))
}

// First defines a new HeapItemValue if heap slot.
func (b *Block) GetPointerToMaybeHeapDefine(store Store, nx *NameExpr) PointerValue {
	switch nx.Type {
	case NameExprTypeNormal:
		// XXX convert rangestmt switchstmt names
		// into NameExpr and then panic here instead.
		return b.GetPointerTo(store, nx.Path)
	case NameExprTypeDefine:
		return b.GetPointerTo(store, nx.Path)
	case NameExprTypeHeapDefine:
		path := nx.Path
		ptr := b.GetPointerToDirect(store, path)
		if _, ok := ptr.TV.T.(heapItemType); ok {
			if nx.Type != NameExprTypeHeapDefine {
				panic("expected name expr heap define type")
			}
			hiv := &HeapItemValue{}
			*ptr.TV = TypedValue{
				T: heapItemType{},
				V: hiv,
			}
			return PointerValue{
				TV:    &hiv.Value,
				Base:  hiv,
				Index: 0,
			}
		} else {
			return ptr
		}
	default:
		panic("unexpected NameExpr type for GetPointerToMaybeHeapDefine")
	}
}

// Result is used has lhs for any assignments to "_".
func (b *Block) GetBlankRef() *TypedValue {
	return &b.Blank
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams1(store Store) (pv1 PointerValue) {
	pv1 = b.GetPointerTo(store, NewValuePathBlock(1, 0, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams2(store Store) (pv1, pv2 PointerValue) {
	pv1 = b.GetPointerTo(store, NewValuePathBlock(1, 0, ""))
	pv2 = b.GetPointerTo(store, NewValuePathBlock(1, 1, ""))
	return
}

// Convenience for implementing nativeBody functions.
func (b *Block) GetParams3(store Store) (pv1, pv2, pv3 PointerValue) {
	pv1 = b.GetPointerTo(store, NewValuePathBlock(1, 0, ""))
	pv2 = b.GetPointerTo(store, NewValuePathBlock(1, 1, ""))
	pv3 = b.GetPointerTo(store, NewValuePathBlock(1, 2, ""))
	return
}

func (b *Block) GetBodyStmt() *bodyStmt {
	return &b.bodyStmt
}

// Used by faux blocks like IfCond and SwitchStmt upon clause match.  e.g.
// source: IfCond, b.Source: IfStmt.  Also used by repl to expand block size
// dynamically. In that case source == b.Source.
// See also PackageNode.PrepareNewValues().
func (b *Block) ExpandWith(alloc *Allocator, source BlockNode) {
	sb := source.GetStaticBlock()
	numNames := int(sb.GetNumNames())
	if len(b.Values) > numNames {
		panic(fmt.Sprintf(
			"unexpected block size shrinkage: %v vs %v",
			len(b.Values), numNames))
	}
	if numNames == len(b.Values) {
		return // nothing to do
	}
	oldNames := len(b.Values)
	newNames := numNames - oldNames
	alloc.AllocateBlockItems(int64(newNames))
	heapItems := source.GetHeapItems()
	bvalues := b.Values
	for i := len(b.Values); i < numNames; i++ {
		tv := sb.Values[i]
		if heapItems[i] {
			bvalues = append(bvalues, TypedValue{
				T: heapItemType{},
				V: alloc.NewHeapItem(tv),
			})
		} else {
			bvalues = append(bvalues, tv)
		}
	}
	b.Values = bvalues
	b.Source = source // otherwise new variables won't show in print or debugger.
}

// RefValue.PkgPath is set if the RefValue refers to a local package or an
// external package and originates from a name by the preprocessor.  In this
// case .ObjectID cannot be set because it may not yet be real.
// NOTE: RefValue Object methods declared in ownership.go
type RefValue struct {
	ObjectID ObjectID  `json:",omitempty"` // If non-zero, PkgPath is empty
	Escaped  bool      `json:",omitempty"` // XXX NOT USED DELETEME
	PkgPath  string    `json:",omitempty"` // If set, ObjectID is non-zero
	Hash     ValueHash `json:",omitempty"` // Set iff not escaped
}

func RefValueFromPackage(pv *PackageValue) RefValue {
	return RefValue{PkgPath: pv.PkgPath}
}

// Returns .ObjectID.
// Does not derive ObjectID from PkgPath if of form RefValue{PkgPath}.
// TODO: consider splitting into another value type to prevent confusion.
func (rv RefValue) GetObjectID() ObjectID {
	if rv.PkgPath != "" {
		panic("unexpected ref value of form RefValue{PkgPath}")
	}
	return rv.ObjectID
}

// Base for a detached singleton (e.g. new(int) or &struct{})
// Conceptually like a Block that holds one value.
// NOTE: It is possible for the value to be external
// while the heap item itself is not; but this
// should not be possible w/ blocks or struct values.
// See test/files/zrealm_crossrealm25a.gno.
// NOTE: could be renamed to HeapItemBaseValue.
// See also note in realm.go about auto-unwrapping.
type HeapItemValue struct {
	ObjectInfo
	Value TypedValue
}

// ----------------------------------------

func defaultStructFields(alloc *Allocator, st *StructType) []TypedValue {
	tvs := alloc.NewStructFields(len(st.Fields))
	for i, ft := range st.Fields {
		if ft.Type.Kind() != InterfaceKind {
			tvs[i] = defaultTypedValue(alloc, ft.Type)
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
		for i := range at.Len {
			tvs[i] = defaultTypedValue(alloc, et)
		}
	}
	return av
}

func defaultTypedValue(alloc *Allocator, t Type) TypedValue {
	switch ct := baseOf(t).(type) {
	case nil:
		panic("unexpected nil type")
	case *InterfaceType:
		return TypedValue{}
	case *ArrayType:
		return TypedValue{
			T: t,
			V: defaultArrayValue(alloc, ct),
		}
	case *StructType:
		return TypedValue{
			T: t,
			V: defaultStructValue(alloc, ct),
		}
	case *SliceType:
		return TypedValue{
			T: t,
			V: nil,
		}
	case *MapType:
		return TypedValue{
			T: t,
			V: nil,
		}
	default:
		return TypedValue{
			T: t,
			V: nil,
		}
	}
}

func typedInt(i int) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(int64(i))
	return tv
}

func typedBool(b bool) TypedValue {
	tv := TypedValue{T: BoolType}
	tv.SetBool(b)
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

// returns the same tv instance for convenience.
func fillValueTV(store Store, tv *TypedValue) *TypedValue {
	switch cv := tv.V.(type) {
	case *HeapItemValue:
		fillValueTV(store, &cv.Value)
	case RefValue:
		if cv.PkgPath != "" { // load package
			tv.V = store.GetPackage(cv.PkgPath, false)
		} else { // load object
			tv.V = store.GetObject(cv.ObjectID)
		}
	case PointerValue:
		// As a special case, cv.Base is filled
		// and cv.TV set appropriately.
		// XXX but why, isn't lazy better?
		// Alternatively, could implement
		// `PointerValue.Deref(store) *TypedValue`,
		// but for execution speed traded off for
		// loading speed, we do the following for now:
		switch cbv := cv.Base.(type) {
		case *HeapItemValue:
			fillValueTV(store, &cbv.Value)
		case RefValue:
			base := store.GetObject(cbv.ObjectID).(Value)
			cv.Base = base
			switch cbv := base.(type) {
			case *ArrayValue:
				et := baseOf(tv.T).(*PointerType).Elt
				epv := cbv.GetPointerAtIndexInt2(store, cv.Index, et)
				cv.TV = epv.TV // TODO optimize? (epv.* ignored)
			case *StructValue:
				fpv := cbv.GetPointerToInt(store, cv.Index)
				cv.TV = fpv.TV // TODO optimize?
			case *BoundMethodValue:
				panic("should not happen")
			case *MapValue:
				panic("should not happen")
			case *Block:
				vpv := cbv.GetPointerToInt(store, cv.Index)
				cv.TV = vpv.TV // TODO optimize?
			case *HeapItemValue:
				cv.TV = &cbv.Value
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

// ----------------------------------------
// Utility
func signOfInteger[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}](v T) int {
	if v > 0 {
		return 1
	} else if v < 0 {
		return -1
	default:
		return 1
	}
}

func signOfFloat32Bits(u32 uint32) int {
	sign, mant, exp, _, nan := softfloat.Funpack32(u32)
	if nan {
		panic("sign of NaN is undefined")
	}
	if exp == 0 && mant == 0 {
		return 0
	}
	if sign != 0 {
		return -1
	}
	return 1
}

func signOfFloat64Bits(u64 uint64) int {
	sign, mant, exp, _, nan := softfloat.Funpack64(u64)
	if nan {
		panic("sign of NaN is undefined")
	}
	if exp == 0 && mant == 0 {
		return 0
	}
	if sign != 0 {
		return -1
	}
	return 1
}
