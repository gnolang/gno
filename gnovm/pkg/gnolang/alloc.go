package gnolang

import (
	"fmt"
	"math/bits"
	"unsafe"

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
	collect  func() (left int64, ok bool) // gc callback
	gasMeter store.GasMeter
}

// for gonative, which doesn't consider the allocator.
var nilAllocator = (*Allocator)(nil)

// Allocation size constants for gas metering.
//
// Raw sizes (_alloc*) are unsafe.Sizeof for each GnoVM value type.
// These must be updated when struct fields change.
// Run `go run misc/devtools/checksize.go` to verify.
//
// Composite sizes (alloc*) represent total heap cost:
//
//	_allocHeap: Go runtime per-object overhead (conservative).
//
//	By-pointer types (*StructValue, *FuncValue, etc.) implement
//	Value with pointer receivers. Creating one heap-allocates the
//	struct. Cost: _allocHeap + sizeof.
//
//	By-value types (PointerValue, RefValue, etc.) implement Value
//	with value receivers. Storing in TypedValue.V (an interface)
//	escapes them to heap. Cost: _allocHeap + sizeof.
//
//	BigintValue/BigdecValue are pointer-sized (8 bytes) and don't
//	escape, but their internal *big.Int/*apd.Decimal are heap-
//	allocated. _allocBigint/_allocBigdec estimate that cost.
//
//	Variable-size components (string bytes, slice items, struct
//	fields, map items) are counted separately per element.
const (
	_allocHeap = 32 // Go heap allocation overhead (conservative)

	// By-value types (value receivers on Value interface).
	// Escape to heap when stored in TypedValue.V.
	_allocPointerValue = 32 // unsafe.Sizeof(PointerValue{})
	_allocRefValue     = 72 // unsafe.Sizeof(RefValue{})
	_allocTypeValue    = 16 // unsafe.Sizeof(TypeValue{})
	_allocTypedValue   = 40 // unsafe.Sizeof(TypedValue{})

	// By-pointer types (pointer receivers on Value interface).
	// Heap-allocated; *T stored in TypedValue.V.
	_allocStructValue      = 176 // unsafe.Sizeof(StructValue{})
	_allocArrayValue       = 200 // unsafe.Sizeof(ArrayValue{})
	_allocSliceValue       = 40  // unsafe.Sizeof(SliceValue{})
	_allocFuncValue        = 352 // unsafe.Sizeof(FuncValue{})
	_allocMapValue         = 168 // unsafe.Sizeof(MapValue{})
	_allocBoundMethodValue = 200 // unsafe.Sizeof(BoundMethodValue{})
	_allocBlock            = 528 // unsafe.Sizeof(Block{})
	_allocPackageValue     = 272 // unsafe.Sizeof(PackageValue{})
	_allocHeapItemValue    = 192 // unsafe.Sizeof(HeapItemValue{})
	_allocRefNode          = 88  // unsafe.Sizeof(RefNode{}) -- TODO verify

	// Estimated heap sizes for pointed-to objects.
	// BigintValue and BigdecValue are just 8-byte pointers;
	// these estimate the *big.Int / *apd.Decimal internals.
	_allocBigint = 200 // estimated: big.Int + typical nat slice
	_allocBigdec = 200 // estimated: apd.Decimal + internals
	_allocType   = 200 // estimated: average Type implementation
	_allocAny    = 200 // estimated: generic fallback

)

const (
	// StringValue is a Go string (16 bytes, by value).
	// Bytes are counted separately via allocStringByte.
	allocString     = _allocHeap + 16
	allocStringByte = 1

	// BigintValue (8 bytes, fits in interface word, no escape).
	// Cost is the internal *big.Int heap object.
	allocBigint     = _allocHeap + _allocBigint
	allocBigintByte = 1

	// BigdecValue (8 bytes, fits in interface word, no escape).
	// Cost is the internal *apd.Decimal heap object.
	allocBigdec     = _allocHeap + _allocBigdec
	allocBigdecByte = 1

	// PointerValue (32 bytes, by value, escapes to heap via interface).
	allocPointer = _allocHeap + _allocPointerValue

	// By-pointer types: _allocHeap + sizeof.
	allocArray       = _allocHeap + _allocArrayValue
	allocArrayItem   = _allocTypedValue
	allocSlice       = _allocHeap + _allocSliceValue
	allocStruct      = _allocHeap + _allocStructValue
	allocStructField = _allocTypedValue
	allocFunc        = _allocHeap + _allocFuncValue
	allocMap         = _allocHeap + _allocMapValue
	allocMapItem     = _allocTypedValue * 2 // key + value TypedValues
	allocBoundMethod = _allocHeap + _allocBoundMethodValue
	allocBlock       = _allocHeap + _allocBlock
	allocBlockItem   = _allocTypedValue
	allocHeapItem    = _allocHeap + _allocHeapItemValue
	allocPackage     = _allocHeap + _allocPackageValue

	// RefValue (72 bytes, by value, escapes to heap via interface).
	allocRefValue = _allocHeap + _allocRefValue
	// RefNode (88 bytes, by value).
	allocRefNode = _allocHeap + _allocRefNode

	// Type is an interface; implementations vary.
	allocType = _allocHeap + _allocType

	allocDataByte   = 1
	allocTypedValue = _allocTypedValue
)

func init() {
	check := func(name string, constant uintptr, actual uintptr) {
		if constant != actual {
			panic("alloc constant " + name + " is stale; update to match unsafe.Sizeof")
		}
	}
	check("_allocPointerValue", _allocPointerValue, unsafe.Sizeof(PointerValue{}))
	check("_allocStructValue", _allocStructValue, unsafe.Sizeof(StructValue{}))
	check("_allocArrayValue", _allocArrayValue, unsafe.Sizeof(ArrayValue{}))
	check("_allocSliceValue", _allocSliceValue, unsafe.Sizeof(SliceValue{}))
	check("_allocFuncValue", _allocFuncValue, unsafe.Sizeof(FuncValue{}))
	check("_allocMapValue", _allocMapValue, unsafe.Sizeof(MapValue{}))
	check("_allocBoundMethodValue", _allocBoundMethodValue, unsafe.Sizeof(BoundMethodValue{}))
	check("_allocBlock", _allocBlock, unsafe.Sizeof(Block{}))
	check("_allocPackageValue", _allocPackageValue, unsafe.Sizeof(PackageValue{}))
	check("_allocTypeValue", _allocTypeValue, unsafe.Sizeof(TypeValue{}))
	check("_allocTypedValue", _allocTypedValue, unsafe.Sizeof(TypedValue{}))
	check("_allocRefValue", _allocRefValue, unsafe.Sizeof(RefValue{}))
	check("_allocHeapItemValue", _allocHeapItemValue, unsafe.Sizeof(HeapItemValue{}))
}

// allocGasTable[k] = gas for a Go heap allocation of 2^k bytes.
// 1 gas = 1 nanosecond on reference hardware.
// Calibrated from Go 1.24 / linux / amd64 benchmarks on DigitalOcean Dedicated (2-core).
// CPU: Intel Xeon Platinum 8168 @ 2.70GHz.
//
// Model: entries [0]-[5] (1B-32B) are exact benchmark medians. Entries [6]-[30]
// use a power-law fit: ns = 0.47 × size^0.925 (straight line in log-log space).
// At runtime, allocGas uses bits.Len64 + linear interpolation (O(1), ~1.5ns).
//
// See gnovm/cmd/calibrate/ for benchmarks, data, and regeneration instructions.
var allocGasTable = [32]int64{
	12,        // 2^0  =       1B   (12ns)
	13,        // 2^1  =       2B   (13ns)
	13,        // 2^2  =       4B   (13ns)
	15,        // 2^3  =       8B   (15ns)
	23,        // 2^4  =      16B   (23ns)
	27,        // 2^5  =      32B   (27ns)
	36,        // 2^6  =      64B   (36ns)
	52,        // 2^7  =     128B   (52ns)
	82,        // 2^8  =     256B   (82ns)
	145,       // 2^9  =     512B   (145ns)
	241,       // 2^10 =      1KB   (241ns)
	458,       // 2^11 =      2KB   (458ns)
	848,       // 2^12 =      4KB   (848ns)
	1637,      // 2^13 =      8KB   (1637ns)
	2798,      // 2^14 =     16KB   (2798ns)
	5520,      // 2^15 =     32KB   (5520ns)
	10611,     // 2^16 =     64KB   (10611ns)
	23513,     // 2^17 =    128KB   (23513ns)
	49114,     // 2^18 =    256KB   (49114ns)
	75155,     // 2^19 =    512KB   (75155ns)
	195342,    // 2^20 =      1MB   (195342ns)
	195342,    // 2^21 =      2MB   (bench: 171176ns, clamped to entry[20] for monotonicity)
	275629,    // 2^22 =      4MB   (275629ns)
	1188110,   // 2^23 =      8MB   (1188110ns)
	2544343,   // 2^24 =     16MB   (2544343ns)
	4987722,   // 2^25 =     32MB   (4987722ns)
	9789111,   // 2^26 =     64MB   (9789111ns)
	19398830,  // 2^27 =    128MB   (19398830ns)
	38412058,  // 2^28 =    256MB   (38412058ns)
	76146343,  // 2^29 =    512MB   (76146343ns)
	153376629, // 2^30 =      1GB   (153376629ns)
	291415595, // 2^31 =      2GB   (power-law extrapolation: ~1.9x from 1GB)
}

// allocGas returns the gas cost for a heap allocation of the given size
// in bytes. Uses a lookup table indexed by floor(log2(size)) with linear
// interpolation, giving O(1) cost (~1.5ns via CLZ instruction).
func allocGas(size int64) int64 {
	if size <= 1 {
		return allocGasTable[0]
	}
	k := bits.Len64(uint64(size)) - 1 // floor(log2(size))
	if k >= 31 {
		return allocGasTable[31]
	}
	lo := allocGasTable[k]
	hi := allocGasTable[k+1]
	frac := size - (int64(1) << k)
	span := int64(1) << k
	return lo + (hi-lo)*frac/span
}

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

// Recount adds size to bytes without charging gas.
// Used during GC re-walk to re-count surviving objects
// without double-charging for already-paid allocations.
func (alloc *Allocator) Recount(size int64) {
	alloc.bytes += size
}

// Fork creates a new Allocator with the same limits but no gasMeter
// or GC callback. The caller must set these via SetGasMeter/SetGCFn
// if gas charging or GC is needed (e.g. for transactions).
// Query contexts intentionally omit the gasMeter.
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
		if alloc.collect == nil {
			// Forked allocators (e.g. the store's tx-scoped allocator
			// before NewMachineWithOptions installs a GC callback, and
			// query-path store allocators which never get one) have no
			// collect function — there's nothing to GC, so cap is final.
			panic("allocation limit exceeded (no GC)")
		}
		if left, ok := alloc.collect(); !ok {
			// GC could not free enough.
			panic("allocation limit exceeded")
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

	// Charge allocation gas based on calibrated lookup table.
	// Models actual CPU time of Go's malloc + zero-fill.
	if alloc.gasMeter != nil {
		alloc.gasMeter.ConsumeGas(allocGas(size), "memory allocation")
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

	return allocPackage
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
		ss += allocRefNode
	}

	ss += allocBlock + allocBlockItem*int64(len(b.Values))

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

// internalRefSize computes the allocation size for RefValue slots stored
// within an object's amino-serialized form. During copyValueWithRefs, child
// Objects are replaced with RefValue{ObjectID, Hash} placeholders. These
// RefValue slots are not counted by GetShallowSize (which only counts the
// object's own structure), so we must account for them here to keep the
// allocator consistent with the store's memory usage.
func internalRefSize(val Value) int64 {
	var size int64
	switch v := val.(type) {
	case *PackageValue:
		if _, ok := v.Block.(RefValue); ok {
			size += allocRefValue // .Block ref
		}
		// include RefValue size
		for _, fb := range v.FBlocks {
			if _, ok := fb.(RefValue); !ok {
				continue
			}
			size += allocRefValue
		}
	case *Block:
		for _, v := range v.Values {
			if _, ok := v.V.(RefValue); ok {
				size += allocRefValue
			}
		}
		if _, ok := v.Parent.(RefValue); ok {
			size += allocRefValue
		}
	case *ArrayValue:
		if v.Data == nil {
			for _, tv := range v.List {
				if _, ok := tv.V.(RefValue); ok {
					size += allocRefValue
				}
			}
		}
	case *StructValue:
		for _, tv := range v.Fields {
			if _, ok := tv.V.(RefValue); ok {
				size += allocRefValue
			}
		}
	case *MapValue:
		for cur := v.List.Head; cur != nil; cur = cur.Next {
			if _, ok := cur.Key.V.(RefValue); ok {
				size += allocRefValue
			}
			if _, ok := cur.Value.V.(RefValue); ok {
				size += allocRefValue
			}
		}
	case *BoundMethodValue:
		if _, ok := v.Receiver.V.(RefValue); ok {
			size += allocRefValue
		}
	case *HeapItemValue:
		if _, ok := v.Value.V.(RefValue); ok {
			size += allocRefValue
		}
	case RefValue:
		// do nothing
	case *PointerValue:
		if _, ok := v.Base.(RefValue); ok {
			size += allocRefValue
		}
	case *SliceValue:
		if _, ok := v.Base.(RefValue); ok {
			size += allocRefValue
		}
	case *FuncValue:
		for _, tv := range v.Captures {
			if _, ok := tv.V.(RefValue); ok {
				size += allocRefValue
			}
		}
		if _, ok := v.Parent.(RefValue); ok {
			size += allocRefValue
		}
	case StringValue:
		// do nothing
	case BigintValue:
		// do nothing
	case BigdecValue:
		// do nothing
	case DataByteValue:
		// do nothing
	case TypeValue:
		// do nothing
	default:
		panic(fmt.Sprintf(
			"unexpected type %T",
			val))
	}
	return size
}
