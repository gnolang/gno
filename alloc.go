package gno

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	bytes int64
}

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
	//allocNative = 1
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

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	alloc.Allocate(+allocStructField * fields)
}

func (alloc *Allocator) AllocateByteArray(size int64) {
	alloc.Allocate(allocArray + size)
}

func (alloc *Allocator) AllocateItemArray(items int64) {
	alloc.Allocate(allocArray + allocArrayItem*items)
}

func (alloc *Allocator) AllocateString(size int64) {
	alloc.Allocate(allocString + allocStringByte*size)
}

//----------------------------------------
// constructor utilities.

func (alloc *Allocator) NewStringValue(s string) StringValue {
	alloc.AllocateString(int64(len(s)))
	return StringValue(s)
}
