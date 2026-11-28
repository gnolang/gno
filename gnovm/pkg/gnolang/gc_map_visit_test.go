package gnolang

import (
	"fmt"
	"math"
	"testing"
)

// gcTestMap builds an N-entry MapValue whose entries come from mk, linked in
// insertion order as MapValue.Assign would.
func gcTestMap(n int, mk func(i int) (k, v TypedValue)) *MapValue {
	mv := &MapValue{List: &MapList{}}
	mv.vmap = make(map[MapKey]*MapListItem, n)
	for i := 0; i < n; i++ {
		k, v := mk(i)
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

func gcTestInt(i int) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(int64(i))
	return tv
}

// gcTestVisits runs the production GC visitor over v and returns its count.
func gcTestVisits(v Value) int64 {
	var visitCount int64
	vis := GCVisitorFn(1, NewAllocator(math.MaxInt64), &visitCount)
	vis(v)
	return visitCount
}

// Regression for gnolang/gno#5883: every MapListItem the visitor walks must
// be counted, whatever the entry holds. Before the fix only boxed keys and
// values were counted (and only on first reach), so an N-entry map of
// primitives — or of N references to one shared object — cost ~1 visit for
// an O(N) traversal.
func TestGCMapVisitCount(t *testing.T) {
	const N = 1000
	sharedInner := gcTestMap(1, func(i int) (TypedValue, TypedValue) { return gcTestInt(i), gcTestInt(i) })
	innerT := &MapType{Key: IntType, Value: IntType}
	sharedFn := &FuncValue{}
	str := func(s string) TypedValue { return TypedValue{T: StringType, V: StringValue(s)} }

	cases := []struct {
		name string
		mv   *MapValue
		want int64
	}{
		{
			// map + N entries.
			"map[int]int",
			gcTestMap(N, func(i int) (TypedValue, TypedValue) { return gcTestInt(i), gcTestInt(i) }),
			1 + N,
		},
		{
			// map + N entries + N string values (strings are not Objects,
			// so each is visited).
			"map[int]string",
			gcTestMap(N, func(i int) (TypedValue, TypedValue) { return gcTestInt(i), str("v") }),
			1 + N + N,
		},
		{
			// map + N entries + N string keys.
			"map[string]int",
			gcTestMap(N, func(i int) (TypedValue, TypedValue) { return str(fmt.Sprint(i)), gcTestInt(i) }),
			1 + N + N,
		},
		{
			// map + N entries + shared inner map once (1 + its 1 entry).
			"map[int]map[int]int shared inner",
			gcTestMap(N, func(i int) (TypedValue, TypedValue) {
				return gcTestInt(i), TypedValue{T: innerT, V: sharedInner}
			}),
			1 + N + 2,
		},
		{
			// map + N entries + shared func once.
			"map[int]func() shared",
			gcTestMap(N, func(i int) (TypedValue, TypedValue) {
				return gcTestInt(i), TypedValue{T: &FuncType{}, V: sharedFn}
			}),
			1 + N + 1,
		},
		{
			"empty map",
			gcTestMap(0, nil),
			1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Fresh cycle stamps: the shared objects are reused across cases.
			sharedInner.SetLastGCCycle(0)
			sharedFn.SetLastGCCycle(0)
			if got := gcTestVisits(tc.mv); got != tc.want {
				t.Fatalf("visitCount = %d, want %d", got, tc.want)
			}
		})
	}
}

// A map is visited once per cycle however many references reach it: its
// entries must not be re-counted on later reaches.
func TestGCMapVisitCountedOncePerCycle(t *testing.T) {
	const N = 500
	inner := gcTestMap(N, func(i int) (TypedValue, TypedValue) { return gcTestInt(i), gcTestInt(i) })
	innerT := &MapType{Key: IntType, Value: IntType}
	av := &ArrayValue{List: make([]TypedValue, 10)}
	for i := range av.List {
		av.List[i] = TypedValue{T: innerT, V: inner}
	}
	// array + inner map + its N entries.
	if got, want := gcTestVisits(av), int64(1+1+N); got != want {
		t.Fatalf("visitCount = %d, want %d", got, want)
	}
}

// End-to-end through GarbageCollect on real Gno code: GC gas for a package
// holding an O(N) map must grow with N for every entry shape, including the
// shared-value shape that bypassed the first version of this fix.
func TestGCMapGasScalesWithEntries(t *testing.T) {
	shapes := map[string]string{
		"primitive": `var M = map[int]int{}
func main() { for i := 0; i < %d; i++ { M[i] = i } }`,
		"shared inner map": `var M = map[int]map[int]int{}
func main() { in := map[int]int{1: 1}; for i := 0; i < %d; i++ { M[i] = in } }`,
		"shared func": `var M = map[int]func(){}
func f() {}
func main() { for i := 0; i < %d; i++ { M[i] = f } }`,
		"map in struct in slice": `type S struct{ M map[int]bool }
var X []S
func main() { s := S{M: map[int]bool{}}; for i := 0; i < %d; i++ { s.M[i] = true }; X = append(X, s) }`,
		"interface keys": `var M = map[interface{}]int{}
func main() { for i := 0; i < %d; i++ { M[i] = i } }`,
	}
	gcGas := func(t *testing.T, body string) int64 {
		t.Helper()
		m := newOperandStackTestMachine(t, 1<<32, true)
		m.RunFiles(m.MustParseFile("main.gno", "package gcstack\n"+body))
		m.RunMain()
		before := m.GasMeter.GasConsumed()
		if _, ok := m.GarbageCollect(); !ok {
			t.Fatal("GarbageCollect failed")
		}
		return m.GasMeter.GasConsumed() - before
	}
	const N = 5000
	for name, src := range shapes {
		t.Run(name, func(t *testing.T) {
			small := gcGas(t, fmt.Sprintf(src, 10))
			large := gcGas(t, fmt.Sprintf(src, N))
			// Each extra entry is at least one visit at the cheapest rate.
			minDelta := int64(N-10) * gcVisitGasTable[0]
			if large-small < minDelta {
				t.Fatalf("GC gas %d (N=10) -> %d (N=%d): delta %d < %d; map walk not metered",
					small, large, N, large-small, minDelta)
			}
		})
	}
}
