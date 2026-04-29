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

// TestStringGCRecount verifies string byte counting behavior across GC cycles:
//  1. Within one GC cycle, shared backings (s1 := s) are counted only once.
//  2. Across GC cycles, the full string bytes are recounted each cycle.
//  3. Dead strings (not visited) are cleaned up after GC.
func TestStringGCRecount(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	// Create a tracked string via NewString.
	sv := alloc.NewString("hello world, this is a test string")
	strLen := int64(len(sv))

	// Verify it's tracked.
	if _, exists := alloc.allocStrings[uintptr(unsafe.Pointer(unsafe.StringData(string(sv))))]; !exists {
		t.Fatal("NewString did not track the backing pointer")
	}

	// --- GC cycle 1 ---
	gcCycle1 := int64(1)
	var vc1 int64
	vis1 := GCVisitorFn(gcCycle1, alloc, &vc1)

	alloc.Reset()

	// First visit: should count full string bytes.
	vis1(sv)
	bytesAfterFirst := alloc.bytes
	headerSize := int64(allocString)
	expectedFull := headerSize + allocStringByte*strLen
	if bytesAfterFirst != expectedFull {
		t.Errorf("cycle 1, first visit: got %d bytes, want %d (header %d + %d bytes)",
			bytesAfterFirst, expectedFull, headerSize, strLen)
	}

	// Second visit (simulating s1 := s, shared backing): header only.
	vis1(sv)
	bytesAfterSecond := alloc.bytes
	wantAfterSecond := expectedFull + headerSize // +headerSize: second visit counts header only (dedup)
	if bytesAfterSecond != wantAfterSecond {
		t.Errorf("cycle 1, second visit: got %d bytes, want %d (previous %d + header %d)",
			bytesAfterSecond, wantAfterSecond, expectedFull, headerSize)
	}

	// Cleanup: visited entry should survive.
	alloc.CleanupTrackedStrings(gcCycle1)
	if len(alloc.allocStrings) != 1 {
		t.Errorf("after cycle 1 cleanup: want 1 tracked entry, got %d", len(alloc.allocStrings))
	}

	// --- GC cycle 2 ---
	gcCycle2 := int64(2)
	var vc2 int64
	vis2 := GCVisitorFn(gcCycle2, alloc, &vc2)

	alloc.Reset()

	// First visit in cycle 2: should count full string bytes again.
	vis2(sv)
	bytesAfterCycle2 := alloc.bytes
	if bytesAfterCycle2 != expectedFull {
		t.Errorf("cycle 2, first visit: got %d bytes, want %d (header %d + %d bytes)",
			bytesAfterCycle2, expectedFull, headerSize, strLen)
	}

	// Cleanup: entry should still survive (visited in cycle 2).
	alloc.CleanupTrackedStrings(gcCycle2)
	if len(alloc.allocStrings) != 1 {
		t.Errorf("after cycle 2 cleanup: want 1 tracked entry, got %d", len(alloc.allocStrings))
	}

	// --- Dead string cleanup ---
	// Simulate a GC cycle where the string is NOT visited.
	gcCycle3 := int64(3)
	alloc.CleanupTrackedStrings(gcCycle3)

	// Entry should be removed (not visited in cycle 3).
	if len(alloc.allocStrings) != 0 {
		t.Errorf("after cycle 3 cleanup (not visited): want 0 tracked entries, got %d", len(alloc.allocStrings))
	}
}

// TestStringSliceGCRecount verifies that a sliced string (s2 := s[x:y])
// counts only the header during GC — the backing bytes are shared with
// the source string and accounted via its allocStrings entry.
func TestStringSliceGCRecount(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	// Create a tracked source string via NewString.
	src := alloc.NewString("abcdefghijklmnopqrstuvwxyz")

	// Simulate s2 := src[2:5] ("cde").
	// Go string slicing shares the backing — only a new header is allocated.
	sliced := StringValue(string(src)[2:5])

	// Verify the sliced string is NOT tracked (only the source is).
	srcPtr := uintptr(unsafe.Pointer(unsafe.StringData(string(src))))
	slicedPtr := uintptr(unsafe.Pointer(unsafe.StringData(string(sliced))))
	if _, ok := alloc.allocStrings[slicedPtr]; ok {
		t.Error("sliced string should NOT be tracked (not created via NewString)")
	}
	t.Logf("source pointer: %d, sliced pointer: %d (offset by 2)", srcPtr, slicedPtr)

	gcCycle := int64(1)
	var vc int64
	vis := GCVisitorFn(gcCycle, alloc, &vc)

	alloc.Reset()

	// Visit source: counts header + full backing bytes.
	vis(src)
	bytesAfterSrc := alloc.bytes
	fullSize := int64(allocString) + allocStringByte*int64(len(src))
	if bytesAfterSrc != fullSize {
		t.Errorf("source visit: got %d, want %d (header + full bytes)", bytesAfterSrc, fullSize)
	}

	// Visit sliced: counts header only (backing shared with source, not tracked).
	vis(sliced)
	bytesAfterSliced := alloc.bytes
	wantAfterSliced := fullSize + int64(allocString) // +header only
	if bytesAfterSliced != wantAfterSliced {
		t.Errorf("sliced visit: got %d, want %d (source + header only for slice)",
			bytesAfterSliced, wantAfterSliced)
	}

	// Source entry should survive cleanup (visited).
	alloc.CleanupTrackedStrings(gcCycle)
	if len(alloc.allocStrings) != 1 {
		t.Errorf("after cleanup: want 1 tracked entry (source only), got %d", len(alloc.allocStrings))
	}
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
