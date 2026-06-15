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

// TestNewMapHintBoundary pins the maxMapHint pivot: hints up to and
// including the constant charge the full preallocation cost; strictly
// above the pivot, plus negatives, silently clamp to 0 (charge only the
// map header). Each case uses its own MaxInt64-budget allocator because
// the pivot case consumes nearly the entire budget.
func TestNewMapHintBoundary(t *testing.T) {
	t.Parallel()

	mt := &MapType{Key: IntType, Value: IntType}

	cases := []struct {
		name string
		size int
		want int64
	}{
		{"pivot-1", maxMapHint - 1, int64(allocMap + allocMapItem*(maxMapHint-1))},
		{"pivot", maxMapHint, int64(allocMap + allocMapItem*maxMapHint)},
		{"pivot+1", maxMapHint + 1, allocMap},
		{"MaxInt", math.MaxInt, allocMap},
		{"neg", -1, allocMap},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			alloc := NewAllocator(math.MaxInt64)
			alloc.NewMap(mt, tc.size)
			if _, b := alloc.Status(); b != tc.want {
				t.Errorf("bytes=%d, want=%d", b, tc.want)
			}
		})
	}
}
