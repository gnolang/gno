package gnolang

import (
	"reflect"
)

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	maxBytes  int64
	bytes     int64
	lastGcRun int64
	minGcRun  int64
	heap      *Heap
	detonate  bool
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
	_allocFuncValue        = 136
	_allocMapValue         = 144
	_allocBoundMethodValue = 176
	_allocBlock            = 464
	_allocNativeValue      = 48
	_allocTypeValue        = 16
	_allocTypedValue       = 40
	_allocBigint           = 200 // XXX
	_allocBigdec           = 200 // XXX
	_allocType             = 200 // XXX
	_allocAny              = 200 // XXX
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
	allocNative      = _allocBase + _allocPointer + _allocNativeValue
	allocType        = _allocBase + _allocPointer + _allocType
	// allocDataByte    = 1
	// allocPackge = 1
	allocAmino     = _allocBase + _allocPointer + _allocAny
	allocAminoByte = 10 // XXX
	allocHeapItem  = _allocBase + _allocPointer + _allocTypedValue
)

func NewAllocator(maxBytes int64, minGcRun int64, heap *Heap) *Allocator {
	if maxBytes == 0 {
		return nil
	}
	return &Allocator{
		maxBytes: maxBytes,
		heap:     heap,
		minGcRun: minGcRun,
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

	alloc.bytes += size
	if alloc.bytes > alloc.maxBytes {
		if alloc.heap != nil {
			deleted := alloc.heap.MarkAndSweep()
			alloc.DeallocDeleted(deleted)
			if alloc.detonate && alloc.bytes > alloc.maxBytes {
				panic("allocation limit exceeded")
			}
			alloc.detonate = alloc.bytes > alloc.maxBytes
		}
	} else if alloc.ShouldRunGC() {
		alloc.lastGcRun = alloc.bytes
		deleted := alloc.heap.MarkAndSweep()
		alloc.DeallocDeleted(deleted)
	}
}

func (alloc *Allocator) ShouldRunGC() bool {
	return alloc.bytes >= 2*alloc.lastGcRun && alloc.bytes >= alloc.minGcRun
}

func (alloc *Allocator) DeallocDeleted(objs []*GcObj) {
	for _, obj := range objs {
		alloc.DeallocObj(obj.tv)
	}
}

func (alloc *Allocator) DeallocObj(tv TypedValue) {
	switch v := tv.V.(type) {
	case PointerValue:
		alloc.DeallocObj(*v.TV)
	case *StructValue:
		alloc.DeallocateStruct()
		alloc.DeallocateStructFields(int64(len(v.Fields)))

		for _, field := range v.Fields {
			alloc.DeallocObj(field)
		}
	case *SliceValue:
		alloc.DeallocateSlice()
	case *ArrayValue:
		alloc.DeallocateDataArray(int64(len(v.Data)))
	default:
		switch tv.T {
		default:
			//fmt.Printf("DeallocDeleted: unimplemented %T\n", v)
		}
	}
}

func (alloc *Allocator) Deallocate(size int64) {
	if alloc == nil {
		// this can happen for map items just prior to assignment.
		return
	}

	alloc.bytes -= size
	if alloc.bytes < 0 {
		panic("Deallocate called with negative size")
	}
}

func (alloc *Allocator) DeallocateString(size int64) {
	alloc.Deallocate(allocString + allocStringByte*size)
}

func (alloc *Allocator) AllocateString(size int64) {
	alloc.Allocate(allocString + allocStringByte*size)
}

func (alloc *Allocator) AllocatePointer() {
	alloc.Allocate(allocPointer)
}

func (alloc *Allocator) AllocateDataArray(size int64) {
	alloc.Allocate(allocArray + size)
}

func (alloc *Allocator) DeallocateDataArray(size int64) {
	alloc.Deallocate(allocArray + size)
}

func (alloc *Allocator) AllocateListArray(items int64) {
	alloc.Allocate(allocArray + allocArrayItem*items)
}

func (alloc *Allocator) AllocateSlice() {
	alloc.Allocate(allocSlice)
}

func (alloc *Allocator) DeallocateSlice() {
	alloc.Deallocate(allocSlice)
}

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) DeallocateStruct() {
	alloc.Deallocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	alloc.Allocate(allocStructField * fields)
}

func (alloc *Allocator) DeallocateStructFields(fields int64) {
	alloc.Deallocate(allocStructField * fields)
}

func (alloc *Allocator) AllocateFunc() {
	alloc.Allocate(allocFunc)
}

func (alloc *Allocator) AllocateMap(items int64) {
	alloc.Allocate(allocMap + allocMapItem*items)
}

func (alloc *Allocator) AllocateMapItem() {
	alloc.Allocate(allocMapItem)
}

func (alloc *Allocator) AllocateBoundMethod() {
	alloc.Allocate(allocBoundMethod)
}

func (alloc *Allocator) AllocateBlock(items int64) {
	alloc.Allocate(allocBlock + allocBlockItem*items)
}

func (alloc *Allocator) DeallocateBlock(items int64) {
	alloc.Deallocate(allocBlock + allocBlockItem*items)
}

func (alloc *Allocator) AllocateBlockItems(items int64) {
	alloc.Allocate(allocBlockItem * items)
}

// NOTE: does not allocate for the underlying value.
func (alloc *Allocator) AllocateNative() {
	alloc.Allocate(allocNative)
}

/* NOTE: Not used, account for with AllocatePointer.
func (alloc *Allocator) AllocateDataByte() {
	alloc.Allocate(allocDataByte)
}
*/

func (alloc *Allocator) AllocateType() {
	alloc.Allocate(allocType)
}

// NOTE: a reasonable max-bounds calculation for simplicity.
func (alloc *Allocator) AllocateAmino(l int64) {
	alloc.Allocate(allocAmino + allocAminoByte*l)
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
	alloc.AllocateListArray(int64(n))
	return &ArrayValue{
		List: make([]TypedValue, n),
	}
}

func (alloc *Allocator) NewDataArray(n int) *ArrayValue {
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

// NOTE: also allocates the underlying array from list.
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

// NOTE: also allocates the underlying array from data.
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

func (alloc *Allocator) NewBlock(source BlockNode, parent *Block) *Block {
	alloc.AllocateBlock(int64(source.GetNumNames()))
	return NewBlock(source, parent)
}

func (alloc *Allocator) NewNative(rv reflect.Value) *NativeValue {
	alloc.AllocateNative()
	return &NativeValue{
		Value: rv,
	}
}

func (alloc *Allocator) NewType(t Type) Type {
	alloc.AllocateType()
	return t
}

func (alloc *Allocator) NewHeapItem(tv TypedValue) *HeapItemValue {
	alloc.AllocateHeapItem()

	if alloc != nil {
		gcObj := NewObject(tv)
		gcObj.marked = true
		alloc.heap.AddObject(gcObj)
	}

	return &HeapItemValue{Value: tv}
}

func (alloc *Allocator) DropPointers(ptrs []PointerValue) {
	if alloc == nil {
		return
	}

	for _, ptr := range ptrs {
		if pv, ok := ptr.TV.V.(PointerValue); ok {
			if _, ok := pv.Base.(*HeapItemValue); ok {
				root := NewObject(*ptr.TV)
				alloc.heap.RemoveRoot(root)
			}
		}
	}
}
