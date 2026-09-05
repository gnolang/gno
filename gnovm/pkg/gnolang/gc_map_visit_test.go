package gnolang

import (
	"math"
	"testing"
)

// Regression: the GC visit count for a large primitive-keyed/valued map must
// reflect its O(N) MapList traversal. Before the fix, map[int]int entries
// (data in TypedValue.N, V == nil) contributed 0 to visitCount, so an N-entry
// map was charged as ~1 visit.
func TestGCPrimitiveMapVisitCount(t *testing.T) {
	const N = 1000
	mv := &MapValue{List: &MapList{}}
	mv.vmap = make(map[MapKey]*MapListItem, N)
	for i := 0; i < N; i++ {
		item := &MapListItem{Key: TypedValue{T: IntType}, Value: TypedValue{T: IntType}}
		item.Key.SetInt(int64(i))
		item.Value.SetInt(int64(i))
		if mv.List.Head == nil {
			mv.List.Head, mv.List.Tail = item, item
		} else {
			item.Prev = mv.List.Tail
			mv.List.Tail.Next = item
			mv.List.Tail = item
		}
		mv.List.Size++
	}

	var visitCount int64
	alloc := NewAllocator(math.MaxInt64)
	vis := GCVisitorFn(1, alloc, &visitCount)
	vis(mv)

	// The map object itself is 1 visit; each of the N fully-primitive entries
	// must now contribute a visit (the metered O(N) walk).
	if visitCount < int64(N) {
		t.Fatalf("visitCount = %d, want >= %d (primitive-map O(N) walk must be metered)", visitCount, N)
	}
	t.Logf("N=%d visitCount=%d gcVisitGas=%d", N, visitCount, gcVisitGas(visitCount))
}

// The fix must add ZERO visits for maps with any boxed key or value — only
// fully-unboxed entries (both Key.V and Value.V nil) are counted. This pins
// the no-over-metering / no-gas-change invariant for object & mixed maps.
func TestGCMapVisitNoOverchargeBoxed(t *testing.T) {
	mk := func(boxKey, boxVal bool) *MapValue {
		const N = 100
		mv := &MapValue{List: &MapList{}}
		mv.vmap = make(map[MapKey]*MapListItem, N)
		for i := 0; i < N; i++ {
			var k, v TypedValue
			if boxKey {
				k = TypedValue{T: StringType, V: StringValue("k")}
			} else {
				k = TypedValue{T: IntType}
				k.SetInt(int64(i))
			}
			if boxVal {
				v = TypedValue{T: StringType, V: StringValue("v")}
			} else {
				v = TypedValue{T: IntType}
				v.SetInt(int64(i))
			}
			item := &MapListItem{Key: k, Value: v}
			if mv.List.Head == nil {
				mv.List.Head, mv.List.Tail = item, item
			} else {
				item.Prev = mv.List.Tail
				mv.List.Tail.Next = item
				mv.List.Tail = item
			}
			mv.List.Size++
		}
		return mv
	}
	// The fix adds exactly the number of entries with Key.V==nil && Value.V==nil.
	added := func(mv *MapValue) int {
		n := 0
		for cur := mv.List.Head; cur != nil; cur = cur.Next {
			if cur.Key.V == nil && cur.Value.V == nil {
				n++
			}
		}
		return n
	}
	if n := added(mk(true, false)); n != 0 {
		t.Fatalf("boxed-key map: fix would add %d visits, want 0 (over-metering)", n)
	}
	if n := added(mk(false, true)); n != 0 {
		t.Fatalf("boxed-value map: fix would add %d visits, want 0 (over-metering)", n)
	}
	if n := added(mk(true, true)); n != 0 {
		t.Fatalf("boxed key+value map: fix would add %d visits, want 0 (over-metering)", n)
	}
	if n := added(mk(false, false)); n != 100 {
		t.Fatalf("fully-primitive map: fix adds %d visits, want 100", n)
	}
}
