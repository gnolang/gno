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
	realmAllocString      = 1
	realmAllocStringByte  = 1
	realmAllocBigint      = 1
	realmAllocBigintByte  = 1
	realmAllocPointer     = 1
	realmAllocArray       = 1
	realmAllocArrayItem   = 1
	realmAllocSlice       = 1
	realmAllocStruct      = 1
	realmAllocFunc        = 1
	realmAllocMap         = 1
	realmAllocMapItem     = 1
	realmAllocBoundMethod = 1
	realmAllocBlock       = 1
	realmAllocBlockItem   = 1
	//realmAllocType = 1
	//realmAllocPackge = 1
	//realmAllocNative = 1
)

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (all *Allocator) Allocate(size int64) {
	all.bytes += size
	if all.bytes > maxAllocations {
		panic("allocation limit exceeded")
	}
}
