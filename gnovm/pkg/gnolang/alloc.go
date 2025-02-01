package gnolang

import (
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	m            *Machine
	maxBytes     int64
	bytes        int64
	throwAway    bool
	StagingBytes int64
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

func NewAllocator(maxBytes int64, m *Machine) *Allocator {
	debug2.Println2("NewAllocator(), maxBytes:", maxBytes)
	if maxBytes == 0 {
		return nil
	}
	return &Allocator{
		maxBytes: maxBytes,
		m:        m,
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

func (alloc *Allocator) MemStats() string {
	if alloc == nil {
		return "nil allocator"
	} else {
		return fmt.Sprintf("Allocator{maxBytes:%d, bytes:%d}", alloc.maxBytes, alloc.bytes)
	}
}

func (alloc *Allocator) GC() {
	gasCPU := overflow.Mul64p(16, GasFactorCPU)
	if alloc.m.GasMeter != nil { //  no gas meter for test
		debug2.Println2("Consuming gas for GC")
		alloc.m.GasMeter.ConsumeGas(gasCPU, "GC")
	}
	fmt.Println("---gc, MemStats:", alloc.MemStats())
	fmt.Println("StagingBytes: ", alloc.StagingBytes)
	defer func() {
		fmt.Println("------------after gc, memStats:", alloc.MemStats())
	}()
	// TODO: why this matters, think this defer is not gc'd
	//defer fmt.Println("------------after gc, memStats:", alloc.MemStats())
	// a throwaway allocator
	throwaway := NewAllocator(alloc.maxBytes, alloc.m)
	throwaway.throwAway = true
	debug2.Println2("m: ", alloc.m)
	debug2.Println2("=====================================================================")

	// scan frames
	for i, fr := range alloc.m.Frames {
		debug2.Printf2("frames[%d]: %v \n", i, fr)

		fv := fr.Func
		debug2.Println2("fv: ", fv)
		if fv != nil {
			ft := fv.GetType(alloc.m.Store)
			if ft.HasVarg() {
				debug2.Println2("has varg")
				pts := ft.Params
				numParams := len(pts)
				isMethod := 0 // 1 if true
				nvar := fr.NumArgs - (numParams - 1 - isMethod)
				throwaway.AllocateSlice()
				throwaway.AllocateListArray(int64(nvar))
			}
		}
		// defer func
		for _, dfr := range fr.Defers {
			fv := dfr.Func
			ft := fv.GetType(alloc.m.Store)
			if ft.HasVarg() {
				debug2.Println2("has defer, has varg")
				numParams := len(ft.Params)
				numArgs := len(dfr.Args)
				nvar := numArgs - (numParams - 1)
				throwaway.AllocateSlice()
				throwaway.AllocateListArray(int64(nvar))
			}
		}
	}

	// scan blocks
	for i, b := range alloc.m.Blocks {
		debug2.Printf2("allocate blocks[%d]: %v \n", i, b)
		throwaway.allocateValue(b)

		// TODO: move this to allocateValue
		for _, tv := range b.Values {
			debug2.Println2("allocate values of block")
			throwaway.allocateTV(tv)
		}
	}

	// reset allocator
	debug2.Println2("---throwaway.bytes: ", throwaway.bytes)
	debug2.Println2("---before reset, alloc.bytes: ", alloc.bytes)
	alloc.bytes = throwaway.bytes + alloc.StagingBytes
	debug2.Println2("---after reset, alloc.bytes: ", alloc.bytes)
}

func (throwaway *Allocator) allocateTV(tv TypedValue) {
	debug2.Println2("allocateTV, tv: ", tv)
	debug2.Println2("allocInfo: ", tv.AllocationInfo.String())

	if tv.GetRefCount() < 2 {
		// allocate type
		if tv.AllocType {
			throwaway.AllocateType()
		}
		if tv.AllocValue {
			throwaway.allocateValue(tv.V)
		}
	} else {
		debug2.Println2("escaped, do nothing, allocInfo: ", tv.AllocationInfo.String())
	}
}

func (throwaway *Allocator) allocateValue(v Value) {
	debug2.Println2("allocateValue: ", v, reflect.TypeOf(v))
	// alloc amino
	// if ref value, allocate amino
	if oo, ok := v.(Object); ok {
		if oid := oo.GetObjectID(); !oid.IsZero() {
			debug2.Println2("oid: ", oid)
			debug2.Println2("oo: ", oo)
			throwaway.AllocateAmino(int64(oo.GetByteSize()))
		} else {
			debug2.Println2("oid: ", oid)
		}
	}
	switch vv := v.(type) {
	case TypeValue: // for newType, not for const type
		debug2.Println2("TypeVal, vv.Type: ", vv.Type, reflect.TypeOf(vv.Type))
		if dt, ok := vv.Type.(*DeclaredType); ok {
			debug2.Println2("TypeVal.Type.(*DeclaredType): ", dt, dt.Base, reflect.TypeOf(dt.Base))
		}
		throwaway.AllocateType()
	case *StructValue:
		// TODO: alloc struct fields first, num of items

		// alloc zero value, first alloc outer struct
		// see var s = S{name: "foo"}
		// the struct lit is allocated first,
		// then with copy, it is allocated again.
		throwaway.AllocateStructFields(int64(len(vv.Fields)))
		throwaway.AllocateStruct()

		// last, alloc fields recursively
		for _, field := range vv.Fields {
			debug2.Println2("alloc field: ", field)
			throwaway.allocateTV(field)
		}
	case *FuncValue:
		throwaway.AllocateFunc()
		//// TODO: is this right?
		//// if closure if fileNode, no allocate,
		//// cuz it's already done in compile stage.
		//debug2.Println2("funcValue, vv: ", vv)
		////debug2.Println2("clo...Source: ", vv.Closure.(*Block).GetSource(throwaway.m.Store), reflect.TypeOf(vv.Closure.(*Block).GetSource(throwaway.m.Store)))
		//if _, ok := vv.Closure.(*Block).GetSource(throwaway.m.Store).(*FileNode); !ok {
		//	debug2.Println2("not global, alloc func")
		//	throwaway.AllocateFunc()
		//} else {
		//	debug2.Println2("global function, NO alloc for it")
		//}
	case PointerValue:
		throwaway.AllocatePointer()
		throwaway.allocateValue(vv.Base)
	case *HeapItemValue:
		throwaway.AllocateHeapItem()
		throwaway.allocateValue(vv.Value.V)
	case *SliceValue:
		throwaway.AllocateSlice()
		throwaway.allocateValue(vv.Base)
	case *ArrayValue:
		// TODO: data array
		if len(vv.List) != 0 {
			throwaway.AllocateListArray(int64(len(vv.List)))
		} else {
			throwaway.AllocateDataArray(int64(len(vv.Data)))
		}
	case *Block:
		throwaway.AllocateBlock(int64(vv.Source.GetNumNames()))
	case StringValue:
		throwaway.AllocateString(int64(len(vv)))
	case *MapValue:
		throwaway.AllocateMap(int64(vv.List.Size))
	case *BoundMethodValue:
		throwaway.AllocateBoundMethod()
	case *NativeValue:
		throwaway.AllocateNative()
	default:
		//debug2.Println2("---default, do nothing: ", vv)
	}
}

func (alloc *Allocator) Allocate(size int64) {
	debug2.Println2("Allocate, size: ", size)
	if alloc == nil {
		debug2.Println2("allocator is nil, do nothing")
		// this can happen for map items just prior to assignment.
		return
	}

	debug2.Println2("alloc.StagingBytes: ", alloc.StagingBytes)
	alloc.StagingBytes += size

	// if alloc on throwaway still exceeds memory,
	// means GC does not work, panic.
	if alloc.throwAway {
		//fmt.Println("throwaway, Allocate, memStats: ", alloc.MemStats())
		if alloc.bytes > alloc.maxBytes {
			debug2.Println2("---exceed memory size............")
			panic("Memory size limit exceeded")
		}
	}

	debug2.Println2("new allocated: ", size)
	alloc.bytes += size
	debug2.Println2("===========bytes after allocated============: ", alloc.bytes)

	if !alloc.throwAway && alloc.bytes > alloc.maxBytes {
		debug2.Println2("---exceed memory size, GC")
		alloc.GC()
		//panic("allocation limit exceeded")
	}
}

func (alloc *Allocator) AllocateString(size int64) {
	debug2.Println2("AllocateString, size: ", size)
	alloc.Allocate(allocString + allocStringByte*size)
}

func (alloc *Allocator) AllocatePointer() {
	debug2.Println2("AllocatePointer")
	alloc.Allocate(allocPointer)
}

func (alloc *Allocator) AllocateDataArray(size int64) {
	debug2.Println2("AllocateDataArray")
	alloc.Allocate(allocArray + size)
}

func (alloc *Allocator) AllocateListArray(items int64) {
	debug2.Println2("AllocateListArray, items: ", items)
	alloc.Allocate(allocArray + allocArrayItem*items)
}

func (alloc *Allocator) AllocateSlice() {
	debug2.Println2("AllocateSlice")
	alloc.Allocate(allocSlice)
}

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	debug2.Println2("AllocateStruct")
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	debug2.Println2("AllocateStructFields")
	alloc.Allocate(allocStructField * fields)
}

func (alloc *Allocator) AllocateFunc() {
	debug2.Println2("AllocateFunc")
	alloc.Allocate(allocFunc)
}

func (alloc *Allocator) AllocateMap(items int64) {
	debug2.Println2("AllocateMap, items: ", items)
	alloc.Allocate(allocMap + allocMapItem*items)
}

func (alloc *Allocator) AllocateMapItem() {
	debug2.Println2("AllocateMapItem")
	alloc.Allocate(allocMapItem)
}

func (alloc *Allocator) AllocateBoundMethod() {
	debug2.Println2("AllocateBoundMethod")
	alloc.Allocate(allocBoundMethod)
}

func (alloc *Allocator) AllocateBlock(items int64) {
	debug2.Println2("AllocateBlock, items: ", items)
	alloc.Allocate(allocBlock + allocBlockItem*items)
}

func (alloc *Allocator) AllocateBlockItems(items int64) {
	debug2.Println2("AllocateBlockItems, items: ", items)
	alloc.Allocate(allocBlockItem * items)
}

// NOTE: does not allocate for the underlying value.
func (alloc *Allocator) AllocateNative() {
	debug2.Println2("AllocateNative")
	alloc.Allocate(allocNative)
}

/* NOTE: Not used, account for with AllocatePointer.
func (alloc *Allocator) AllocateDataByte() {
	alloc.Allocate(allocDataByte)
}
*/

func (alloc *Allocator) AllocateType() {
	debug2.Println2("AllocateType")
	alloc.Allocate(allocType)
}

// NOTE: a reasonable max-bounds calculation for simplicity.
func (alloc *Allocator) AllocateAmino(l int64) {
	debug2.Println2("AllocateAmino, l: ", l)
	alloc.Allocate(allocAmino + allocAminoByte*l)
}

func (alloc *Allocator) AllocateHeapItem() {
	debug2.Println2("AllocateHeapItem")
	alloc.Allocate(allocHeapItem)
}

//----------------------------------------
// constructor utilities.

func (alloc *Allocator) NewString(s string) StringValue {
	debug2.Printf2("NewString, s: \"%s\" \n", s)
	alloc.AllocateString(int64(len(s)))
	return StringValue(s)
}

func (alloc *Allocator) NewListArray(n int) *ArrayValue {
	debug2.Println2("NewListArray: ", n)
	if n < 0 {
		panic(&Exception{Value: typedString("len out of range")})
	}
	alloc.AllocateListArray(int64(n))
	return &ArrayValue{
		List: make([]TypedValue, n),
	}
}

func (alloc *Allocator) NewDataArray(n int) *ArrayValue {
	debug2.Println2("NewDataArray: ", n)
	alloc.m.LastBlock()
	if n < 0 {
		panic(&Exception{Value: typedString("len out of range")})
	}

	alloc.AllocateDataArray(int64(n))
	return &ArrayValue{
		Data: make([]byte, n),
	}
}

func (alloc *Allocator) NewArrayFromData(data []byte) *ArrayValue {
	debug2.Println2("NewArrayFromData: ", len(data))
	av := alloc.NewDataArray(len(data))
	copy(av.Data, data)
	return av
}

func (alloc *Allocator) NewSlice(base Value, offset, length, maxcap int) *SliceValue {
	debug2.Println2("NewSlice: ", base)
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
func (alloc *Allocator) NewSliceFromList(list []TypedValue) *SliceValue {
	debug2.Println2("NewSliceFromList: ", len(list))
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
	debug2.Println2("NewSliceFromData: ", len(data))
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
	debug2.Println2("NewStruct: ", fields)
	alloc.AllocateStruct()
	return &StructValue{
		Fields: fields,
	}
}

func (alloc *Allocator) NewStructFields(fields int) []TypedValue {
	debug2.Println2("NewStructFields", fields)
	alloc.AllocateStructFields(int64(fields))
	return make([]TypedValue, fields)
}

// NOTE: fields will be allocated.
func (alloc *Allocator) NewStructWithFields(fields ...TypedValue) *StructValue {
	debug2.Println2("NewStructWithFields", fields)
	tvs := alloc.NewStructFields(len(fields))
	copy(tvs, fields)
	return alloc.NewStruct(tvs)
}

func (alloc *Allocator) NewMap(size int) *MapValue {
	debug2.Println2("NewMap, size: ", size)
	alloc.AllocateMap(int64(size))
	mv := &MapValue{}
	mv.MakeMap(size)
	return mv
}

func (alloc *Allocator) NewBlock(source BlockNode, parent *Block) *Block {
	debug2.Printf2("NewBlock, source: %v, source...Names: %v\n", source, source.GetBlockNames())
	alloc.AllocateBlock(int64(source.GetNumNames()))
	return NewBlock(source, parent)
}

func (alloc *Allocator) NewNative(rv reflect.Value) *NativeValue {
	debug2.Println2("NewNative", rv)
	alloc.AllocateNative()
	return &NativeValue{
		Value: rv,
	}
}

func (alloc *Allocator) NewType(t Type) Type {
	debug2.Println2("NewType:", t)
	alloc.AllocateType()
	return t
}

func (alloc *Allocator) NewHeapItem(tv TypedValue) *HeapItemValue {
	debug2.Println2("NewHeapItem", tv)
	alloc.AllocateHeapItem()
	return &HeapItemValue{Value: tv}
}

// ========================================================
type AllocationInfo struct {
	AllocType  bool
	AllocValue bool
	RefCount   int
}

func (ai *AllocationInfo) String() string {
	return fmt.Sprintf(
		"AllocationInfo{AllocType: %t, AllocValue: %t, RefCount: %d}",
		ai.AllocType, ai.AllocValue, ai.RefCount,
	)
}

func (ai *AllocationInfo) SetAllocValue(allocValue bool) {
	(*ai).AllocValue = allocValue
}

func (ai *AllocationInfo) SetAllocType(allocType bool) {
	ai.AllocType = allocType
}

func (ai *AllocationInfo) IncRefCount() {
	ai.RefCount++
}

func (ai *AllocationInfo) GetRefCount() int {
	return ai.RefCount
}
