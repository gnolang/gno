package gnolang

import (
	"math"
	"slices"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
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

// rangeFor returns the index of the range containing p, or -1.
func (alloc *Allocator) rangeFor(p uintptr) int {
	for i, r := range alloc.stringRanges {
		if p >= r.start && p < r.end {
			return i
		}
	}
	return -1
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
	srcPtr := uintptr(unsafe.Pointer(unsafe.StringData(string(sv))))
	if alloc.rangeFor(srcPtr) < 0 {
		t.Fatal("NewString did not register a range covering the backing pointer")
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
	if len(alloc.stringRanges) != 1 {
		t.Errorf("after cycle 1 cleanup: want 1 tracked range, got %d", len(alloc.stringRanges))
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
	if len(alloc.stringRanges) != 1 {
		t.Errorf("after cycle 2 cleanup: want 1 tracked range, got %d", len(alloc.stringRanges))
	}

	// --- Dead string cleanup ---
	// Simulate a GC cycle where the string is NOT visited.
	gcCycle3 := int64(3)
	alloc.CleanupTrackedStrings(gcCycle3)

	// Entry should be removed (not visited in cycle 3).
	if len(alloc.stringRanges) != 0 {
		t.Errorf("after cycle 3 cleanup (not visited): want 0 tracked ranges, got %d", len(alloc.stringRanges))
	}
}

// TestStringSliceGCRecount verifies that a sliced string (s2 := s[x:y])
// resolves to the source's range via containment — no new range is added
// for the slice itself, and the visitor charges the source's full backing
// bytes only on the first visit per cycle.
func TestStringSliceGCRecount(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	src := alloc.NewString("abcdefghijklmnopqrstuvwxyz")

	// Simulate s2 := src[2:5] ("cde"). Go shares the backing; only header alloc.
	sliced := StringValue(string(src)[2:5])

	// Slice's ptr resolves into the source's range via containment.
	srcPtr := uintptr(unsafe.Pointer(unsafe.StringData(string(src))))
	slicedPtr := uintptr(unsafe.Pointer(unsafe.StringData(string(sliced))))
	srcIdx := alloc.rangeFor(srcPtr)
	slicedIdx := alloc.rangeFor(slicedPtr)
	if srcIdx < 0 || slicedIdx < 0 {
		t.Fatalf("expected both ptrs to resolve; src=%d sliced=%d", srcIdx, slicedIdx)
	}
	if srcIdx != slicedIdx {
		t.Errorf("source and slice should resolve to the same range, got %d vs %d", srcIdx, slicedIdx)
	}
	if got := len(alloc.stringRanges); got != 1 {
		t.Errorf("expected 1 range (source only), got %d", got)
	}

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

	// Visit sliced: header only (range already counted this cycle).
	vis(sliced)
	bytesAfterSliced := alloc.bytes
	wantAfterSliced := fullSize + int64(allocString)
	if bytesAfterSliced != wantAfterSliced {
		t.Errorf("sliced visit: got %d, want %d (source + header only for slice)",
			bytesAfterSliced, wantAfterSliced)
	}

	alloc.CleanupTrackedStrings(gcCycle)
	if len(alloc.stringRanges) != 1 {
		t.Errorf("after cleanup: want 1 tracked range (source), got %d", len(alloc.stringRanges))
	}
}

// TestStringSliceOutlivesSource is the regression test for the bug
// thehowl flagged on values.go:2191: when the source string dies but a
// slice with offset M>0 stays alive, the slice's backing must still be
// counted. With uintptr-equality keying, the slice's ptr (src+M) was
// never a key in the map, so its bytes silently disappeared from the
// budget after the source's entry was cleaned up. Range-by-containment
// fixes this — the slice's pointer resolves into the source's range.
func TestStringSliceOutlivesSource(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	src := alloc.NewString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 30 bytes
	srcLen := int64(len(src))
	sliced := StringValue(string(src)[1:]) // ptr = src+1, len 29

	gcCycle := int64(1)
	var vc int64
	vis := GCVisitorFn(gcCycle, alloc, &vc)

	alloc.Reset()

	// Source out of GC roots: only the slice is visited.
	vis(sliced)
	got := alloc.bytes
	want := int64(allocString) + allocStringByte*srcLen
	if got != want {
		t.Errorf("slice-only visit: got %d, want %d (header + full source backing %d)",
			got, want, srcLen)
	}

	// The source's range was refreshed by the slice's lookup, so cleanup keeps it.
	alloc.CleanupTrackedStrings(gcCycle)
	if len(alloc.stringRanges) != 1 {
		t.Errorf("after cleanup: want 1 tracked range, got %d", len(alloc.stringRanges))
	}

	// Next cycle: slice is still alive, range still resolves, bytes recharged.
	gcCycle2 := int64(2)
	vis2 := GCVisitorFn(gcCycle2, alloc, &vc)
	alloc.Reset()
	vis2(sliced)
	if alloc.bytes != want {
		t.Errorf("cycle 2 slice-only visit: got %d, want %d", alloc.bytes, want)
	}
}

// TestFillTypesOfValue_StringTracking verifies the load-path contract:
// when a persisted StringValue is rehydrated through fillTypesOfValue,
// its backing must be registered in the rehydrating allocator's
// tracking structure. Without this, a string that pre-existed the
// current tx allocator would never be tracked, and CountStringBytes
// would silently skip its bytes during GC.
func TestFillTypesOfValue_StringTracking(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	st.SetAllocator(NewAllocator(1_000_000))

	const body = "loaded-from-store"
	loaded := fillTypesOfValue(st, StringValue(body))

	sv, ok := loaded.(StringValue)
	if !ok {
		t.Fatalf("fillTypesOfValue returned %T, want StringValue", loaded)
	}
	if string(sv) != body {
		t.Fatalf("fillTypesOfValue mutated content: got %q, want %q", string(sv), body)
	}

	alloc := st.GetAllocator()
	p := uintptr(unsafe.Pointer(unsafe.StringData(string(sv))))
	if alloc.rangeFor(p) < 0 {
		t.Fatal("fillTypesOfValue did not register the string's backing")
	}
}

// TestNewString_EmptyStringNotTracked verifies the len==0 guard:
// empty strings must not enter stringRanges. unsafe.StringData on an
// empty string returns an unspecified (typically shared sentinel)
// pointer, so tracking would alias every empty string onto one entry.
func TestNewString_EmptyStringNotTracked(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	_ = alloc.NewString("")
	_ = alloc.NewString("")
	if got := len(alloc.stringRanges); got != 0 {
		t.Errorf("empty strings should not be tracked, got %d entries", got)
	}

	if size, ok := alloc.CountStringBytes("", 1); ok || size != 0 {
		t.Errorf("CountStringBytes(\"\") = (%d, %v), want (0, false)", size, ok)
	}
}

// TestTrackString_OverlapClones verifies the determinism guarantee: a
// string whose backing overlaps an already-tracked range (toolchain
// sharing — concat returning its operand, copy elision — or NewString on
// a sub-extent) is cloned onto a fresh backing and tracked as its own
// range, so the range set never depends on Go's sharing decisions.
func TestTrackString_OverlapClones(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	src := alloc.NewString("the quick brown fox")
	if got := len(alloc.stringRanges); got != 1 {
		t.Fatalf("after NewString(src): got %d ranges, want 1", got)
	}
	srcStart, srcEnd := stringExtent(string(src))

	// A sub-extent of src overlaps its range: TrackString must clone it
	// onto a fresh backing and register that as a second range.
	sub := string(src)[4:9]
	tracked := alloc.TrackString(sub)
	if tracked != sub {
		t.Errorf("clone changed content: got %q, want %q", tracked, sub)
	}
	if got := len(alloc.stringRanges); got != 2 {
		t.Fatalf("TrackString of an overlapping extent should clone+track, got %d ranges", got)
	}
	p, _ := stringExtent(tracked)
	if p >= srcStart && p < srcEnd {
		t.Error("tracked string still shares src's backing; expected a clone")
	}

	// A non-overlapping string is tracked as-is, no clone.
	fresh := strings.Repeat("z", 8)
	fp, _ := stringExtent(fresh)
	gp, _ := stringExtent(alloc.TrackString(fresh))
	if gp != fp {
		t.Error("non-overlapping string should be tracked without cloning")
	}
	if got := len(alloc.stringRanges); got != 3 {
		t.Errorf("want 3 ranges, got %d", got)
	}
}

// TestTrackString_RecycledAddress simulates Go's runtime recycling a dead
// tracked backing's address for a new string: stale entries are
// synthesized around a fresh string's extent. TrackString cannot
// distinguish this from sharing a live backing, so it must clone
// conservatively and register the clone's exact extent; entries it cannot
// prove stale are left for CleanupTrackedStrings, and disjoint live
// entries must be untouched.
func TestTrackString_RecycledAddress(t *testing.T) {
	alloc := NewAllocator(1_000_000)

	s := strings.Repeat("a", 16)
	p, end := stringExtent(s)

	stale := []stringRange{
		{start: p - 8, end: p + 4},     // overlaps head
		{start: p + 6, end: p + 10},    // contained
		{start: end - 2, end: end + 4}, // overlaps tail
	}
	survivors := []stringRange{
		{start: p - 100, end: p - 90, lastCycle: 7},
		{start: end + 50, end: end + 60, lastCycle: 7},
	}
	alloc.stringRanges = slices.Concat(survivors[:1], stale, survivors[1:])
	sort.Slice(alloc.stringRanges, func(i, j int) bool {
		return alloc.stringRanges[i].start < alloc.stringRanges[j].start
	})

	tracked := alloc.TrackString(s)
	if tracked != s {
		t.Errorf("clone changed content: got %q, want %q", tracked, s)
	}

	// The clone's exact extent is registered as its own range.
	tp, tend := stringExtent(tracked)
	if i := alloc.rangeFor(tp); i < 0 || alloc.stringRanges[i].start != tp || alloc.stringRanges[i].end != tend {
		t.Error("tracked string's extent not registered exactly")
	}
	// Disjoint entries are never evicted.
	if alloc.rangeFor(p-95) < 0 || alloc.rangeFor(end+55) < 0 {
		t.Error("disjoint entries were wrongly evicted")
	}
	// Whatever wasn't provably stale is at most deferred to GC cleanup:
	// after a cycle in which only survivors are visited, exactly they remain.
	alloc.CleanupTrackedStrings(7)
	if got := len(alloc.stringRanges); got != 2 {
		t.Errorf("after cleanup: want 2 survivor ranges, got %d", got)
	}
}

// TestFork_StartsWithEmptyTracking verifies that Fork does not carry
// over the parent's string tracking. The child re-registers every
// string it charges through its own NewString / fillTypesOfValue path,
// so the parent's entries are unnecessary; sharing them would also be
// unsafe (the child's CleanupTrackedStrings would prune the parent, and
// query paths fork on a different goroutine). Related: thehowl's review.
func TestFork_StartsWithEmptyTracking(t *testing.T) {
	parent := NewAllocator(1_000_000)
	parent.NewString("parented")
	if got := len(parent.stringRanges); got != 1 {
		t.Fatalf("parent should have 1 range, got %d", got)
	}

	child := parent.Fork()
	if got := len(child.stringRanges); got != 0 {
		t.Errorf("child should start with empty tracking, got %d", got)
	}

	// The child tracks its own strings independently.
	child.NewString("child-owned")
	if got := len(child.stringRanges); got != 1 {
		t.Errorf("child should track its own string, got %d", got)
	}

	// Child mutations must not touch the parent.
	child.CleanupTrackedStrings(99)
	if got := len(parent.stringRanges); got != 1 {
		t.Errorf("parent's ranges must be independent of the child: got %d, want 1", got)
	}
}

// TestNewMapChargesHeaderOnly pins the map allocation model: creating a
// map charges only the header (allocMap), never a per-hint preallocation
// cost. Items are charged one allocMapItem each at insertion time. This
// is what lets GnoVM ignore the make() size hint safely — there is no
// allocMapItem*hint term to overflow or to double-charge against the
// per-item charge.
func TestNewMapChargesHeaderOnly(t *testing.T) {
	t.Parallel()

	mt := &MapType{Key: IntType, Value: IntType}
	alloc := NewAllocator(math.MaxInt64)

	alloc.NewMap(mt)
	if _, b := alloc.Status(); b != allocMap {
		t.Fatalf("NewMap charged %d bytes, want allocMap=%d", b, allocMap)
	}

	alloc.AllocateMapItem()
	if _, b := alloc.Status(); b != allocMap+allocMapItem {
		t.Fatalf("after one item charged %d bytes, want %d", b, allocMap+allocMapItem)
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
