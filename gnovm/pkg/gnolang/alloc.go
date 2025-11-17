package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	maxBytes int64
	bytes    int64
	// `peakBytes` represents the maximum memory
	// usage during a single transaction, and is used
	// to calculate the corresponding gas usage.
	// It increases monotonically.
	peakBytes int64
	collect   func() (left int64, ok bool) // gc callback
	gasMeter  store.GasMeter
}

// for gonative, which doesn't consider the allocator.
var nilAllocator = (*Allocator)(nil)

const (
	// go elemental
	_allocBase    = 24 // defensive... XXX
	_allocPointer = 8
	// gno types
	_allocSlice            = 24
	_allocPointerValue     = 40
	_allocStructValue      = 152
	_allocArrayValue       = 176
	_allocSliceValue       = 40
	_allocFuncValue        = 312
	_allocMapValue         = 144
	_allocBoundMethodValue = 176
	_allocBlock            = 472
	_allocPackageValue     = 240
	_allocTypeValue        = 16
	_allocTypedValue       = 40
	_allocRefValue         = 72
	_allocRefNode          = 88
	_allocBigint           = 200 // XXX
	_allocBigdec           = 200 // XXX
	_allocType             = 200 // XXX
	_allocAny              = 200 // XXX
	_allocValue            = 16  // interface
	_allocName             = 16  // string
)

const (
	allocString      = _allocBase
	allocStringByte  = 1
	allocBigint      = _allocBase + _allocPointer + _allocBigint
	allocBigintByte  = 1
	allocBigdec      = _allocBase + _allocPointer + _allocBigdec
	allocBigdecByte  = 1
	allocPointer     = _allocBase
	allocArray       = _allocBase + _allocPointer + _allocArrayValue
	allocArrayItem   = _allocTypedValue
	allocSlice       = _allocBase + _allocPointer + _allocSliceValue
	allocStruct      = _allocBase + _allocPointer + _allocStructValue
	allocStructField = _allocTypedValue
	allocFunc        = _allocBase + _allocPointer + _allocFuncValue
	allocMap         = _allocBase + _allocPointer + _allocMapValue
	allocMapItem     = _allocTypedValue * 3 // XXX
	allocBoundMethod = _allocBase + _allocPointer + _allocBoundMethodValue
	allocBlock       = _allocBase + _allocPointer + _allocBlock
	allocBlockItem   = _allocTypedValue
	allocRefValue    = _allocBase + _allocRefValue
	allocRefNode     = _allocBase + _allocRefNode
	allocType        = _allocBase + _allocPointer + _allocType
	allocDataByte    = 1
	allocPackage     = _allocBase + _allocPointer + _allocPackageValue
	allocHeapItem    = _allocBase + _allocPointer + _allocTypedValue
	allocTypedValue  = _allocTypedValue
)

const GasCostPerByte = 1 // gas cost per byte allocated

func NewAllocator(maxBytes int64) *Allocator {
	if maxBytes == 0 {
		return nil
	}
	return &Allocator{
		maxBytes: maxBytes,
	}
}

func (alloc *Allocator) SetGCFn(f func() (int64, bool)) {
	alloc.collect = f
}

func (alloc *Allocator) SetGasMeter(gasMeter store.GasMeter) {
	alloc.gasMeter = gasMeter
}

func (alloc *Allocator) MemStats() string {
	if alloc == nil {
		return "nil allocator"
	} else {
		return fmt.Sprintf("Allocator{maxBytes:%d, bytes:%d}", alloc.maxBytes, alloc.bytes)
	}
}

func (alloc *Allocator) Status() (maxBytes int64, bytes int64) {
	return alloc.maxBytes, alloc.bytes
}

func (alloc *Allocator) Reset() *Allocator {
	if alloc == nil {
		return nil
	}
	alloc.bytes = 0
	return alloc
}

func (alloc *Allocator) Fork() *Allocator {
	if alloc == nil {
		return nil
	}
	return &Allocator{
		maxBytes: alloc.maxBytes,
		bytes:    alloc.bytes,
	}
}

func (alloc *Allocator) Allocate(size int64) {
	if alloc == nil {
		// this can happen for map items just prior to assignment.
		return
	}
	if overflow.Addp(alloc.bytes, size) > alloc.maxBytes {
		if left, ok := alloc.collect(); !ok {
			panic("should not happen, allocation limit exceeded while gc.")
		} else {
			if debug {
				debug.Printf("GC finished, %d left after GC, required size: %d\n", left, size)
			}
			// retry after GC
			alloc.bytes += size
			if alloc.bytes > alloc.maxBytes {
				panic("allocation limit exceeded")
			}
		}
	} else {
		alloc.bytes += size
	}
	// The value of `bytes` decreases during GC, and fees
	// are only charged when it exceeds peakBytes (again).
	if alloc.bytes > alloc.peakBytes {
		if alloc.gasMeter != nil {
			change := alloc.bytes - alloc.peakBytes
			alloc.gasMeter.ConsumeGas(overflow.Mulp(change, GasCostPerByte), "memory allocation")
		}

		alloc.peakBytes = alloc.bytes
	}
}

func (alloc *Allocator) AllocateString(size int64) {
	alloc.Allocate(overflow.Addp(allocString, overflow.Mulp(allocStringByte, size)))
}

func (alloc *Allocator) AllocatePointer() {
	alloc.Allocate(allocPointer)
}

func (alloc *Allocator) AllocateDataArray(size int64) {
	alloc.Allocate(overflow.Addp(allocArray, size))
}

func (alloc *Allocator) AllocateListArray(items int64) {
	alloc.Allocate(overflow.Addp(allocArray, overflow.Mulp(allocArrayItem, items)))
}

func (alloc *Allocator) AllocateSlice() {
	alloc.Allocate(allocSlice)
}

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	alloc.Allocate(overflow.Mulp(allocStructField, fields))
}

func (alloc *Allocator) AllocateFunc() {
	alloc.Allocate(allocFunc)
}

func (alloc *Allocator) AllocateMap(items int64) {
	alloc.Allocate(overflow.Addp(allocMap, overflow.Mulp(allocMapItem, items)))
}

func (alloc *Allocator) AllocateMapItem() {
	alloc.Allocate(allocMapItem)
}

func (alloc *Allocator) AllocateBoundMethod() {
	alloc.Allocate(allocBoundMethod)
}

func (alloc *Allocator) AllocatePackageValue() {
	alloc.Allocate(allocPackage)
}

func (alloc *Allocator) AllocateBlock(items int64) {
	alloc.Allocate(overflow.Addp(allocBlock, overflow.Mulp(allocBlockItem, items)))
}

func (alloc *Allocator) AllocateBlockItems(items int64) {
	alloc.Allocate(overflow.Mulp(allocBlockItem, items))
}

/* NOTE: Not used, account for with AllocatePointer.
func (alloc *Allocator) AllocateDataByte() {
	alloc.Allocate(allocDataByte)
}
*/

func (alloc *Allocator) AllocateType() {
	alloc.Allocate(allocType)
}

func (alloc *Allocator) AllocateHeapItem() {
	alloc.Allocate(allocHeapItem)
}

//----------------------------------------
// constructor utilities.

func (alloc *Allocator) NewString(s string) StringValue {
	alloc.AllocateString(int64(len(s)))
	return StringValue(s)
}

func (alloc *Allocator) NewListArray(n int) *ArrayValue {
	if n < 0 {
		panic(&Exception{Value: typedString("len out of range")})
	}
	alloc.AllocateListArray(int64(n))
	return &ArrayValue{
		List: make([]TypedValue, n),
	}
}

func (alloc *Allocator) NewListArray2(l, c int) *ArrayValue {
	if l < 0 || c < 0 {
		panic(&Exception{Value: typedString("len or cap out of range")})
	}

	if c < l {
		panic(&Exception{Value: typedString("length and capacity swapped")})
	}

	alloc.AllocateListArray(int64(c))
	return &ArrayValue{
		List: make([]TypedValue, l, c),
	}
}

func (alloc *Allocator) NewDataArray(n int) *ArrayValue {
	if n < 0 {
		panic(&Exception{Value: typedString("len out of range")})
	}

	alloc.AllocateDataArray(int64(n))
	return &ArrayValue{
		Data: make([]byte, n),
	}
}

func (alloc *Allocator) NewArrayFromData(data []byte) *ArrayValue {
	av := alloc.NewDataArray(len(data))
	copy(av.Data, data)
	return av
}

func (alloc *Allocator) NewSlice(base Value, offset, length, maxcap int) *SliceValue {
	alloc.AllocateSlice()
	return &SliceValue{
		Base:   base,
		Offset: offset,
		Length: length,
		Maxcap: maxcap,
	}
}

// NewSliceFromList allocates a new slice with the underlying array value
// populated from `list`. This should not be called from areas in the codebase
// that are doing allocations with potentially large user provided values, e.g.
// `make()` and `append()`. Using `Alloc.NewListArray` can be used is most cases
// to allocate the space for the `TypedValue` list before doing the allocation
// in the go runtime -- see the `make()` code in uverse.go.
// NOTE: cap(list) is propagated directly into the Gno SliceValue.Maxcap.
// Callers must ensure cap(list) == len(list) to produce deterministic results
// across Go versions (Go's append growth strategy is unspecified).
func (alloc *Allocator) NewSliceFromList(list []TypedValue) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateListArray(int64(cap(list)))
	fullList := list[:cap(list)]
	return &SliceValue{
		Base: &ArrayValue{
			List: fullList,
		},
		Offset: 0,
		Length: len(list),
		Maxcap: cap(list),
	}
}

// NewSliceFromData allocates a new slice with the underlying data array
// value populated from `data`. See the doc for `NewSliceFromList` for
// correct usage notes.
func (alloc *Allocator) NewSliceFromData(data []byte) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateDataArray(int64(cap(data)))
	fullData := data[:cap(data)]
	return &SliceValue{
		Base: &ArrayValue{
			Data: fullData,
		},
		Offset: 0,
		Length: len(data),
		Maxcap: cap(data),
	}
}

// NOTE: fields must be allocated (e.g. from NewStructFields)
func (alloc *Allocator) NewStruct(fields []TypedValue) *StructValue {
	alloc.AllocateStruct()
	return &StructValue{
		Fields: fields,
	}
}

func (alloc *Allocator) NewStructFields(fields int) []TypedValue {
	alloc.AllocateStructFields(int64(fields))
	return make([]TypedValue, fields)
}

// NOTE: fields will be allocated.
func (alloc *Allocator) NewStructWithFields(fields ...TypedValue) *StructValue {
	tvs := alloc.NewStructFields(len(fields))
	copy(tvs, fields)
	return alloc.NewStruct(tvs)
}

func (alloc *Allocator) NewMap(size int) *MapValue {
	alloc.AllocateMap(int64(size))
	mv := &MapValue{}
	mv.MakeMap(size)
	return mv
}

// Only used for constructing the main package
func (alloc *Allocator) NewPackageValue(pn *PackageNode) *PackageValue {
	alloc.AllocatePackageValue()
	alloc.AllocateBlock(int64(pn.GetNumNames()))
	pv := &PackageValue{
		Block: &Block{
			Source: pn,
		},
		PkgName:    pn.PkgName,
		PkgPath:    pn.PkgPath,
		FNames:     nil,
		FBlocks:    nil,
		fBlocksMap: make(map[string]*Block),
	}

	return pv
}

func (alloc *Allocator) NewBlock(source BlockNode, parent *Block) *Block {
	alloc.AllocateBlock(int64(source.GetNumNames()))
	return NewBlock(alloc, source, parent)
}

func (alloc *Allocator) NewType(t Type) Type {
	alloc.AllocateType()
	return t
}

func (alloc *Allocator) NewHeapItem(tv TypedValue) *HeapItemValue {
	alloc.AllocateHeapItem()
	return &HeapItemValue{Value: tv}
}

// -----------------------------------------------
// Utilities for obtaining shallow size

func (pv *PackageValue) GetShallowSize() int64 {
	// .uverse is preloaded
	if pv.PkgPath == ".uverse" {
		return 0
	}

	// Account for the PkgName field in addition to the base package allocation
	// PkgName is of type Name (which is string), so use string allocation pattern
	return allocPackage + allocString + allocStringByte*int64(len(pv.PkgName))
}

func (b *Block) GetShallowSize() int64 {
	// .uverse is preloaded, its descendants will also
	// be skipped.
	if pn, ok := b.Source.(*PackageNode); ok {
		if pn.PkgPath == ".uverse" {
			return 0
		}
	}

	var ss int64
	// RefNode is not value, put it here
	// for convinence
	if _, ok := b.Source.(RefNode); ok {
		ss += allocRefValue
	}

	ss = allocBlock + allocBlockItem*int64(len(b.Values))

	return ss
}

func (av *ArrayValue) GetShallowSize() int64 {
	if av.Data != nil {
		return allocArray + int64(len(av.Data))
	} else {
		return allocArray + int64(len(av.List)*allocArrayItem)
	}
}

func (sv *StructValue) GetShallowSize() int64 {
	return allocStruct + int64(len(sv.Fields))*allocStructField
}

func (mv *MapValue) GetShallowSize() int64 {
	return allocMap + allocMapItem*int64(mv.GetLength())
}

func (bmv *BoundMethodValue) GetShallowSize() int64 {
	// skip .uverse
	if bmv.Func.PkgPath == ".uverse" {
		return 0
	}
	return allocBoundMethod
}

func (hiv *HeapItemValue) GetShallowSize() int64 {
	return allocHeapItem
}

func (rv RefValue) GetShallowSize() int64 {
	return allocRefValue
}

func (pv PointerValue) GetShallowSize() int64 {
	return allocPointer
}

func (sv *SliceValue) GetShallowSize() int64 {
	return allocSlice
}

func (fv *FuncValue) GetShallowSize() int64 {
	if fv.PkgPath == ".uverse" {
		return 0
	}

	var ss int64
	ss = allocFunc
	// RefNode is not value, put it here
	// for convinence
	if _, ok := fv.Source.(RefNode); ok {
		ss += allocRefNode
	}

	return ss
}

func (sv StringValue) GetShallowSize() int64 {
	return allocString + allocStringByte*int64(len(sv))
}

func (biv BigintValue) GetShallowSize() int64 {
	return allocBigint
}

func (bdv BigdecValue) GetShallowSize() int64 {
	return allocBigdec
}

func (dbv DataByteValue) GetShallowSize() int64 {
	return allocDataByte
}

// Do not count during recalculation,
// as the type should  pre-exist.
func (tv TypeValue) GetShallowSize() int64 {
	return 0
}
