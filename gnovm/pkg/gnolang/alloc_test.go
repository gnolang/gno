package gnolang

import (
	"testing"
	"unsafe"
)

func TestAllocSizes(t *testing.T) {
	t.Parallel()

	// go elemental
	println("_allocPointer", unsafe.Sizeof(&StructValue{}))
	println("_allocSlice", unsafe.Sizeof([]byte("12345678901234567890123456789012345678901234567890")))
	// gno types
	println("PointerValue{}", unsafe.Sizeof(PointerValue{}))
	println("StructValue{}", unsafe.Sizeof(StructValue{}))
	println("ArrayValue{}", unsafe.Sizeof(ArrayValue{}))
	println("SliceValue{}", unsafe.Sizeof(SliceValue{}))
	println("FuncValue{}", unsafe.Sizeof(FuncValue{}))
	println("MapValue{}", unsafe.Sizeof(MapValue{}))
	println("BoundMethodValue{}", unsafe.Sizeof(BoundMethodValue{}))
	println("Block{}", unsafe.Sizeof(Block{}))
	println("TypeValue{}", unsafe.Sizeof(TypeValue{}))
	println("TypedValue{}", unsafe.Sizeof(TypedValue{}))
	println("ObjectInfo{}", unsafe.Sizeof(ObjectInfo{}))
}

func TestBlockGetShallowSize_WithRefNodeSource(t *testing.T) {
	t.Parallel()

	const numValues = 5
	normalBlock := &Block{
		Source: &FuncDecl{},
		Values: make([]TypedValue, numValues),
	}
	refNodeBlock := &Block{
		Source: RefNode{Location: Location{PkgPath: "gno.land/r/test/foo"}},
		Values: make([]TypedValue, numValues),
	}

	normalSize := normalBlock.GetShallowSize()
	refNodeSize := refNodeBlock.GetShallowSize()

	expectedRefNodeSize := normalSize + allocRefNode
	if refNodeSize != expectedRefNodeSize {
		t.Errorf("Block with RefNode source: GetShallowSize() = %d, want %d (normal %d + allocRefNode %d)",
			refNodeSize, expectedRefNodeSize, normalSize, allocRefNode)
	}
}

func TestAllocatorNumAllocs(t *testing.T) {
	t.Parallel()

	alloc := NewAllocator(1024)
	if got := alloc.NumAllocs(); got != 0 {
		t.Fatalf("new allocator NumAllocs() = %d, want 0", got)
	}

	alloc.Allocate(10)
	alloc.Allocate(20)
	if got := alloc.NumAllocs(); got != 2 {
		t.Fatalf("after two allocations NumAllocs() = %d, want 2", got)
	}
	if got := alloc.TotalAllocBytes(); got != 30 {
		t.Fatalf("after two allocations TotalAllocBytes() = %d, want 30", got)
	}

	_, bytes := alloc.Status()
	if bytes != 30 {
		t.Fatalf("after two allocations bytes = %d, want 30", bytes)
	}
}

func TestAllocatorNumAllocsResetAndRecount(t *testing.T) {
	t.Parallel()

	alloc := NewAllocator(1024)
	alloc.Allocate(10)
	alloc.Recount(20)
	if got := alloc.NumAllocs(); got != 1 {
		t.Fatalf("after recount NumAllocs() = %d, want 1", got)
	}
	if got := alloc.TotalAllocBytes(); got != 10 {
		t.Fatalf("after recount TotalAllocBytes() = %d, want 10", got)
	}

	_, bytes := alloc.Status()
	if bytes != 30 {
		t.Fatalf("after recount bytes = %d, want 30", bytes)
	}

	alloc.Reset()
	if got := alloc.NumAllocs(); got != 0 {
		t.Fatalf("after reset NumAllocs() = %d, want 0", got)
	}
	if got := alloc.TotalAllocBytes(); got != 0 {
		t.Fatalf("after reset TotalAllocBytes() = %d, want 0", got)
	}

	_, bytes = alloc.Status()
	if bytes != 0 {
		t.Fatalf("after reset bytes = %d, want 0", bytes)
	}
}

func TestAllocatorNumAllocsFork(t *testing.T) {
	t.Parallel()

	alloc := NewAllocator(1024)
	alloc.Allocate(10)

	fork := alloc.Fork()
	if got := fork.NumAllocs(); got != 1 {
		t.Fatalf("fork NumAllocs() = %d, want 1", got)
	}
	if got := fork.TotalAllocBytes(); got != 10 {
		t.Fatalf("fork TotalAllocBytes() = %d, want 10", got)
	}

	_, bytes := fork.Status()
	if bytes != 10 {
		t.Fatalf("fork bytes = %d, want 10", bytes)
	}
}

func TestAllocator_NumAllocsSurvivesGC(t *testing.T) {
	t.Parallel()

	alloc := NewAllocator(1024)
	alloc.Allocate(100)
	alloc.Allocate(200)

	preCount := alloc.NumAllocs()
	if preCount != 2 {
		t.Fatalf("precondition: NumAllocs() = %d, want 2", preCount)
	}

	// expect 100 bytes worth of allocations survived the walk
	alloc.resetLiveBytesForGC()
	alloc.Recount(100)

	postCount := alloc.NumAllocs()
	if postCount < preCount {
		t.Errorf("NumAllocs() did not survive GC: pre=%d post=%d ",
			preCount, postCount)
	}
}

func TestAllocator_TotalAllocBytesSurvivesGC(t *testing.T) {
	t.Parallel()

	alloc := NewAllocator(1024)
	alloc.Allocate(100)
	alloc.Allocate(200)

	preBytes := alloc.TotalAllocBytes()
	if preBytes != 300 {
		t.Fatalf("precondition: TotalAllocBytes() = %d, want 300", preBytes)
	}

	// expect 100 bytes survive, 200 die.
	alloc.resetLiveBytesForGC()
	alloc.Recount(100)

	postBytes := alloc.TotalAllocBytes()
	if postBytes < preBytes {
		t.Errorf("TotalAllocBytes() did not survive GC: pre=%d post=%d",
			preBytes, postBytes)
	}
}
