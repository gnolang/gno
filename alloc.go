package gno

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	bytes int64
}

// for gonative, which doesn't consider the allocator.
var nilAllocator = (*Allocator)(nil)

const maxAllocations = 1000000000 // TODO parameterize. 1GB for now.

const (
	allocString      = 1
	allocStringByte  = 1
	allocBigint      = 1
	allocBigintByte  = 1
	allocPointer     = 1
	allocArray       = 1
	allocArrayItem   = 1
	allocSlice       = 1
	allocStruct      = 1
	allocStructField = 1
	allocFunc        = 1
	allocMap         = 1
	allocMapItem     = 1
	allocBoundMethod = 1
	allocBlock       = 1
	allocBlockItem   = 1
	//allocType = 1
	//allocPackge = 1
	allocNative   = 1
	allocDataByte = 1
	allocType     = 1
)

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (alloc *Allocator) Allocate(size int64) {
	if alloc == nil {
		// this can happen for map items just prior to assignment.
		return
	}
	alloc.bytes += size
	if alloc.bytes > maxAllocations {
		panic("allocation limit exceeded")
	}
}

func (alloc *Allocator) AllocateString(size int64) {
	alloc.Allocate(allocString + allocStringByte*size)
}

func (alloc *Allocator) AllocatePointer() {
	alloc.Allocate(allocPointer)
}

func (alloc *Allocator) AllocateByteArray(size int64) {
	alloc.Allocate(allocArray + size)
}

func (alloc *Allocator) AllocateItemArray(items int64) {
	alloc.Allocate(allocArray + allocArrayItem*items)
}

func (alloc *Allocator) AllocateSlice() {
	alloc.Allocate(allocSlice)
}

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	alloc.Allocate(allocStructField * fields)
}

func (alloc *Allocator) AllocateFunc() {
	alloc.Allocate(allocFunc)
}

func (alloc *Allocator) AllocateMap(items int64) {
	alloc.Allocate(allocMap + allocMapItem*items)
}

func (alloc *Allocator) AllocateBlock(items int64) {
	alloc.Allocate(allocBlock + allocBlockItem*items)
}

// NOTE: does not allocate for the underlying value.
func (alloc *Allocator) AllocateNative() {
	alloc.Allocate(allocNative)
}

func (alloc *Allocator) AllocateDataByte() {
	alloc.Allocate(allocDataByte)
}

func (alloc *Allocator) AllocateType() {
	alloc.Allocate(allocType)
}

//----------------------------------------
// constructor utilities.

func (alloc *Allocator) NewStringValue(s string) StringValue {
	alloc.AllocateString(int64(len(s)))
	return StringValue(s)
}

func (alloc *Allocator) NewSliceFromList(list []TypedValue) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateItemArray(int64(cap(list)))
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

func (alloc *Allocator) NewSliceFromData(data []byte) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateByteArray(int64(cap(data)))
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

func (alloc *Allocator) NewBlock(source BlockNode, parent *Block) *Block {
	alloc.AllocateBlock(int64(source.GetNumNames()))
	return NewBlock(source, parent)
}
