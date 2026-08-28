package gnolang

import (
	"math"
	"runtime"
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

// rangeFor returns the tracked range containing p, or nil.
func (alloc *Allocator) rangeFor(p uintptr) *stringRange {
	return alloc.stringRanges.containing(p)
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
	if alloc.rangeFor(srcPtr) == nil {
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
	if alloc.stringRanges.len() != 1 {
		t.Errorf("after cycle 1 cleanup: want 1 tracked range, got %d", alloc.stringRanges.len())
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
	if alloc.stringRanges.len() != 1 {
		t.Errorf("after cycle 2 cleanup: want 1 tracked range, got %d", alloc.stringRanges.len())
	}

	// --- Dead string cleanup ---
	// Simulate a GC cycle where the string is NOT visited.
	gcCycle3 := int64(3)
	alloc.CleanupTrackedStrings(gcCycle3)

	// Entry should be removed (not visited in cycle 3).
	if alloc.stringRanges.len() != 0 {
		t.Errorf("after cycle 3 cleanup (not visited): want 0 tracked ranges, got %d", alloc.stringRanges.len())
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
	srcRange := alloc.rangeFor(srcPtr)
	slicedRange := alloc.rangeFor(slicedPtr)
	if srcRange == nil || slicedRange == nil {
		t.Fatalf("expected both ptrs to resolve; src=%v sliced=%v", srcRange, slicedRange)
	}
	if srcRange != slicedRange {
		t.Errorf("source and slice should resolve to the same range, got %v vs %v", *srcRange, *slicedRange)
	}
	if got := alloc.stringRanges.len(); got != 1 {
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
	if alloc.stringRanges.len() != 1 {
		t.Errorf("after cleanup: want 1 tracked range (source), got %d", alloc.stringRanges.len())
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
	if alloc.stringRanges.len() != 1 {
		t.Errorf("after cleanup: want 1 tracked range, got %d", alloc.stringRanges.len())
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
	loaded := fillTypesOfValue(nil, st, StringValue(body))

	sv, ok := loaded.(StringValue)
	if !ok {
		t.Fatalf("fillTypesOfValue returned %T, want StringValue", loaded)
	}
	if string(sv) != body {
		t.Fatalf("fillTypesOfValue mutated content: got %q, want %q", string(sv), body)
	}

	alloc := st.GetAllocator()
	p := uintptr(unsafe.Pointer(unsafe.StringData(string(sv))))
	if alloc.rangeFor(p) == nil {
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
	if got := alloc.stringRanges.len(); got != 0 {
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
	if got := alloc.stringRanges.len(); got != 1 {
		t.Fatalf("after NewString(src): got %d ranges, want 1", got)
	}
	srcStart, srcEnd := stringExtent(string(src))

	// A sub-extent of src overlaps its range: trackString must clone it
	// onto a fresh backing and register that as a second range.
	sub := string(src)[4:9]
	tracked := alloc.trackString(sub)
	if tracked != sub {
		t.Errorf("clone changed content: got %q, want %q", tracked, sub)
	}
	if got := alloc.stringRanges.len(); got != 2 {
		t.Fatalf("trackString of an overlapping extent should clone+track, got %d ranges", got)
	}
	p, _ := stringExtent(tracked)
	if p >= srcStart && p < srcEnd {
		t.Error("tracked string still shares src's backing; expected a clone")
	}

	// A non-overlapping string is tracked as-is, no clone.
	fresh := strings.Repeat("z", 8)
	fp, _ := stringExtent(fresh)
	gp, _ := stringExtent(alloc.trackString(fresh))
	if gp != fp {
		t.Error("non-overlapping string should be tracked without cloning")
	}
	if got := alloc.stringRanges.len(); got != 3 {
		t.Errorf("want 3 ranges, got %d", got)
	}
}

// TestTrackString_PinnedBackings is the regression test for the
// recycled-address false hit: tracked ranges pin their backings
// (stringRange.str), so Go can never free a tracked backing and recycle
// its address for a new, untracked string (e.g. VM panic text). Before
// pinning, this scenario produced nondeterministic containment hits —
// 1 to 200 per 20000 attempts, varying run to run in one process — each
// charging a dead range's extent to a string the allocator never minted.
// With pinning, hits must be exactly zero, always.
func TestTrackString_PinnedBackings(t *testing.T) {
	alloc := NewAllocator(1 << 30)

	const size = 4096
	const attempts = 10000

	// Track a batch of strings and drop every test-side reference; only
	// the allocator's pins keep the backings alive.
	for i := range 200 {
		alloc.NewString(strings.Repeat(string(rune('a'+i%26)), size))
	}
	runtime.GC()
	runtime.GC()

	// Allocate fresh strings the way panic text is built (plain Go
	// string construction, never minted through NewString). None may
	// resolve into a tracked range: their memory cannot have been
	// recycled from a pinned backing.
	for i := range attempts {
		b := make([]byte, size)
		for j := range b {
			b[j] = byte('A' + (i+j)%26)
		}
		if n, charge := alloc.CountStringBytes(string(b), 1); charge {
			t.Fatalf("untracked string falsely resolved into a tracked range (charged %d bytes): a pinned backing was recycled", n)
		}
		if i%1000 == 0 {
			runtime.GC()
		}
	}
}

// TestClearObjectCache_ClearsStringTracking pins the per-message reset:
// ClearObjectCache runs before each message of a multi-message tx and
// must drop the string tracking along with the byte count — the next
// message's machine restarts GC cycle numbering, so leftover ranges
// would carry stale lastCycle stamps and pin dead backings.
func TestClearObjectCache_ClearsStringTracking(t *testing.T) {
	tm2Store := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	alloc := NewAllocator(1 << 20)
	st.SetAllocator(alloc)

	alloc.NewString("tracked in message 1")
	if got := alloc.stringRanges.len(); got != 1 {
		t.Fatalf("want 1 tracked range before reset, got %d", got)
	}

	st.ClearObjectCache()

	if got := alloc.stringRanges.len(); got != 0 {
		t.Errorf("ClearObjectCache left %d tracked ranges, want 0", got)
	}
	if _, bytes := alloc.Status(); bytes != 0 {
		t.Errorf("ClearObjectCache left %d charged bytes, want 0", bytes)
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
	if got := parent.stringRanges.len(); got != 1 {
		t.Fatalf("parent should have 1 range, got %d", got)
	}

	child := parent.Fork()
	if got := child.stringRanges.len(); got != 0 {
		t.Errorf("child should start with empty tracking, got %d", got)
	}

	// The child tracks its own strings independently.
	child.NewString("child-owned")
	if got := child.stringRanges.len(); got != 1 {
		t.Errorf("child should track its own string, got %d", got)
	}

	// Child mutations must not touch the parent.
	child.CleanupTrackedStrings(99)
	if got := parent.stringRanges.len(); got != 1 {
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
