package gnolang

import (
	"math"
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

// TestNewMapHintBoundary pins the maxMapHint pivot: a hint exactly at the
// constant must NOT clamp (allocator charges the full preallocation cost),
// while pivot+1 must clamp to 0 (allocator charges just the map header).
// Each side uses its own MaxInt64-budget allocator because the pivot side
// consumes nearly the entire budget.
func TestNewMapHintBoundary(t *testing.T) {
	t.Parallel()

	mt := &MapType{Key: IntType, Value: IntType}

	// pivot - 1: no clamp.
	{
		alloc := NewAllocator(math.MaxInt64)
		alloc.NewMap(mt, maxMapHint-1)
		_, bytes := alloc.Status()
		want := int64(allocMap + allocMapItem*(maxMapHint-1))
		if bytes != want {
			t.Errorf("pivot-1: bytes=%d, want=%d (not clamped)", bytes, want)
		}
	}

	// pivot: no clamp (boundary inclusive).
	{
		alloc := NewAllocator(math.MaxInt64)
		alloc.NewMap(mt, maxMapHint)
		_, bytes := alloc.Status()
		want := int64(allocMap + allocMapItem*maxMapHint)
		if bytes != want {
			t.Errorf("pivot: bytes=%d, want=%d (not clamped)", bytes, want)
		}
	}

	// pivot + 1: clamped to 0, charges only allocMap.
	{
		alloc := NewAllocator(math.MaxInt64)
		alloc.NewMap(mt, maxMapHint+1)
		_, bytes := alloc.Status()
		if bytes != allocMap {
			t.Errorf("pivot+1: bytes=%d, want=%d (clamped to 0)", bytes, allocMap)
		}
	}

	// math.MaxInt: clamped (sanity, matches make19.gno).
	{
		alloc := NewAllocator(math.MaxInt64)
		alloc.NewMap(mt, math.MaxInt)
		_, bytes := alloc.Status()
		if bytes != allocMap {
			t.Errorf("MaxInt: bytes=%d, want=%d (clamped to 0)", bytes, allocMap)
		}
	}

	// Negative: clamped to 0.
	{
		alloc := NewAllocator(math.MaxInt64)
		alloc.NewMap(mt, -1)
		_, bytes := alloc.Status()
		if bytes != allocMap {
			t.Errorf("neg: bytes=%d, want=%d (clamped to 0)", bytes, allocMap)
		}
	}
}
