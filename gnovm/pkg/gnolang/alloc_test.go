package gnolang

import (
	"math"
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

// visitStrings runs one GC visitor over vals and returns the recounted
// byte total. Each call is one GC run (fresh dedup set).
func visitStrings(alloc *Allocator, vals ...Value) int64 {
	var vc int64
	vis := GCVisitorFn(1, alloc, &vc)
	alloc.Reset()
	for _, v := range vals {
		vis(v)
	}
	_, bytes := alloc.Status()
	return bytes
}

// TestStringGCRecount: shared backings (s1 := s copies the ID) are
// counted once per GC run; a fresh run recounts in full.
func TestStringGCRecount(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	sv := alloc.NewString("hello world, this is a test string")
	if sv.ID == 0 || sv.Extent != int64(len(sv.Str)) {
		t.Fatalf("NewString: ID=%d Extent=%d, want nonzero ID and Extent=%d", sv.ID, sv.Extent, len(sv.Str))
	}
	full := int64(allocString) + allocStringByte*sv.Extent

	// One value: header + backing.
	if got := visitStrings(alloc, sv); got != full {
		t.Errorf("single visit: got %d, want %d", got, full)
	}
	// s1 := sv shares the ID: second visit is header-only (dedup).
	s1 := sv
	if got, want := visitStrings(alloc, sv, s1), full+int64(allocString); got != want {
		t.Errorf("shared-ID visits: got %d, want %d", got, want)
	}
	// A fresh GC run recounts in full (per-run dedup set, no carryover).
	if got := visitStrings(alloc, sv); got != full {
		t.Errorf("next run: got %d, want %d", got, full)
	}
}

// TestStringSliceGCRecount: GetSlice inherits ID/Extent, so source and
// slice dedup to one full-backing charge.
func TestStringSliceGCRecount(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	src := alloc.NewString("abcdefghijklmnopqrstuvwxyz")
	tv := TypedValue{T: StringType, V: src}
	sliced := tv.GetSlice(alloc, 2, 5).V.(StringValue)

	if sliced.Str != "cde" {
		t.Fatalf("slice content: got %q", sliced.Str)
	}
	if sliced.ID != src.ID || sliced.Extent != src.Extent {
		t.Fatalf("slice identity: got (ID=%d, Extent=%d), want source's (%d, %d)",
			sliced.ID, sliced.Extent, src.ID, src.Extent)
	}
	full := int64(allocString) + allocStringByte*src.Extent
	if got, want := visitStrings(alloc, src, sliced), full+int64(allocString); got != want {
		t.Errorf("source+slice: got %d, want %d (one backing, two headers)", got, want)
	}
}

// TestStringSliceOutlivesSource: with only the slice visited, the full
// source backing is still charged — the slice carries ID and Extent, so
// nothing depends on the source value surviving.
func TestStringSliceOutlivesSource(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	src := alloc.NewString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 30 bytes
	tv := TypedValue{T: StringType, V: src}
	sliced := tv.GetSlice(alloc, 1, 30).V.(StringValue)

	want := int64(allocString) + allocStringByte*src.Extent
	if got := visitStrings(alloc, sliced); got != want {
		t.Errorf("slice-only visit: got %d, want %d (header + full source backing)", got, want)
	}
}

// TestFillTypesOfValue_StringTracking: a loaded StringValue (ID zero on
// the wire) is re-minted through the store's allocator.
func TestFillTypesOfValue_StringTracking(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	st.SetAllocator(NewAllocator(1_000_000))

	const body = "loaded-from-store"
	loaded := fillTypesOfValue(nil, st, StringValue{Str: body})

	sv, ok := loaded.(StringValue)
	if !ok {
		t.Fatalf("fillTypesOfValue returned %T, want StringValue", loaded)
	}
	if sv.Str != body {
		t.Fatalf("content changed: got %q, want %q", sv.Str, body)
	}
	if sv.ID == 0 || sv.Extent != int64(len(body)) {
		t.Errorf("loaded string not minted: ID=%d Extent=%d", sv.ID, sv.Extent)
	}
	if _, bytes := st.GetAllocator().Status(); bytes == 0 {
		t.Error("loaded string was not charged")
	}
}

// TestNewString_Empty: "" carries no backing — no ID, header-only charge.
func TestNewString_Empty(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	sv := alloc.NewString("")
	if sv.ID != 0 || sv.Extent != 0 {
		t.Errorf("empty string: ID=%d Extent=%d, want 0,0", sv.ID, sv.Extent)
	}
}

// TestNewString_DistinctIDs: identity is the mint event, never the Go
// backing. Equal content, toolchain sharing (lit+\"\" returning its
// operand), or address reuse cannot merge two mints.
func TestNewString_DistinctIDs(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	lit := "the quick brown fox"
	a := alloc.NewString(lit)
	b := alloc.NewString(lit + "") // may share lit's backing; irrelevant
	if a.ID == b.ID {
		t.Fatal("two mints share an ID")
	}
	full := 2*int64(allocString) + 2*allocStringByte*int64(len(lit))
	if got := visitStrings(alloc, a, b); got != full {
		t.Errorf("two mints recount: got %d, want %d (both backings charged)", got, full)
	}
}

// TestUntrackedString_HeaderOnly: VM-internal text (panic values) has ID
// zero and recounts as header only — a deterministic undercount. The
// misattribution class the pin design had to defend against (address
// recycling) is unrepresentable here: identity is an ID, not an address.
func TestUntrackedString_HeaderOnly(t *testing.T) {
	alloc := NewAllocator(1_000_000)
	long := StringValue{Str: strings.Repeat("x", 4096)} // e.g. panic text
	if got, want := visitStrings(alloc, long), int64(allocString); got != want {
		t.Errorf("untracked recount: got %d, want header %d", got, want)
	}
}

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
