package gnolang

import (
	"math/rand"
	"testing"
)

// buildGCGraph constructs a connected object graph of approximately nObjects
// objects with a realistic mix of GnoVM value types. The distribution is:
//
//	40% StructValue, 20% ArrayValue, 15% Block, 10% FuncValue,
//	10% HeapItemValue, 5% MapValue.
//
// After creating all objects, it wires them together so that TypedValue.V
// fields point to random other objects, creating a realistic traversal
// pattern for the GC visitor.
func buildGCGraph(nObjects int) Value {
	rng := rand.New(rand.NewSource(42))

	// Determine counts per type.
	nStruct := nObjects * 40 / 100
	nArray := nObjects * 20 / 100
	nBlock := nObjects * 15 / 100
	nFunc := nObjects * 10 / 100
	nHeap := nObjects * 10 / 100
	nMap := nObjects - nStruct - nArray - nBlock - nFunc - nHeap // remainder (~5%)

	// Collect all objects into a flat slice for cross-linking.
	all := make([]Value, 0, nObjects)

	// A minimal BlockNode source for Blocks. FuncDecl embeds StaticBlock
	// which satisfies BlockNode. We just need GetNumNames/GetHeapItems to
	// work; the default zero values are fine (0 names, nil heap items).
	blockSource := &FuncDecl{} // Body not needed for GC visitor
	blockSource.Body = Body{}

	// --- Create StructValues ---
	for i := 0; i < nStruct; i++ {
		nFields := 3 + rng.Intn(3) // 3-5 fields
		sv := &StructValue{
			Fields: make([]TypedValue, nFields),
		}
		// Initialize fields with a simple type.
		for j := range sv.Fields {
			sv.Fields[j].T = IntType
		}
		all = append(all, sv)
	}

	// --- Create ArrayValues ---
	for i := 0; i < nArray; i++ {
		nElems := 5 + rng.Intn(6) // 5-10 elements
		av := &ArrayValue{
			List: make([]TypedValue, nElems),
		}
		for j := range av.List {
			av.List[j].T = IntType
		}
		all = append(all, av)
	}

	// --- Create Blocks ---
	for i := 0; i < nBlock; i++ {
		nVals := 3 + rng.Intn(3) // 3-5 values
		b := &Block{
			Source: blockSource,
			Values: make([]TypedValue, nVals),
		}
		for j := range b.Values {
			b.Values[j].T = IntType
		}
		all = append(all, b)
	}

	// --- Create FuncValues ---
	for i := 0; i < nFunc; i++ {
		nCaptures := 1 + rng.Intn(3) // 1-3 captures
		fv := &FuncValue{
			PkgPath:  "bench",
			Captures: make([]TypedValue, nCaptures),
			Source:   blockSource,
		}
		for j := range fv.Captures {
			fv.Captures[j].T = IntType
		}
		all = append(all, fv)
	}

	// --- Create HeapItemValues ---
	for i := 0; i < nHeap; i++ {
		hiv := &HeapItemValue{
			Value: TypedValue{T: IntType},
		}
		all = append(all, hiv)
	}

	// --- Create MapValues ---
	for i := 0; i < nMap; i++ {
		mv := &MapValue{}
		mv.MakeMap(0)
		nEntries := 2 + rng.Intn(2) // 2-3 entries
		// We need a non-nil allocator for MapList.Append since it
		// calls AllocateMapItem. Use a large-limit allocator.
		tmpAlloc := NewAllocator(1 << 40)
		for j := 0; j < nEntries; j++ {
			item := mv.List.Append(tmpAlloc, TypedValue{T: IntType})
			item.Value = TypedValue{T: IntType}
		}
		all = append(all, mv)
	}

	// --- Wire objects together ---
	// For each object, set some of its TypedValue.V fields to point
	// to random other objects in the graph.
	for _, obj := range all {
		switch v := obj.(type) {
		case *StructValue:
			for j := range v.Fields {
				if rng.Intn(3) == 0 { // ~33% chance of cross-link
					v.Fields[j].V = all[rng.Intn(len(all))]
				}
			}
		case *ArrayValue:
			for j := range v.List {
				if rng.Intn(4) == 0 { // ~25% chance
					v.List[j].V = all[rng.Intn(len(all))]
				}
			}
		case *Block:
			for j := range v.Values {
				if rng.Intn(3) == 0 {
					v.Values[j].V = all[rng.Intn(len(all))]
				}
			}
			// Set parent to a random block.
			if rng.Intn(2) == 0 {
				for tries := 0; tries < 5; tries++ {
					candidate := all[rng.Intn(len(all))]
					if b, ok := candidate.(*Block); ok && b != v {
						v.Parent = b
						break
					}
				}
			}
		case *FuncValue:
			for j := range v.Captures {
				if rng.Intn(2) == 0 {
					v.Captures[j].V = all[rng.Intn(len(all))]
				}
			}
			// Set parent to a random block.
			for tries := 0; tries < 5; tries++ {
				candidate := all[rng.Intn(len(all))]
				if b, ok := candidate.(*Block); ok {
					v.Parent = b
					break
				}
			}
		case *HeapItemValue:
			v.Value.V = all[rng.Intn(len(all))]
		case *MapValue:
			for cur := v.List.Head; cur != nil; cur = cur.Next {
				if rng.Intn(2) == 0 {
					cur.Value.V = all[rng.Intn(len(all))]
				}
			}
		}
	}

	// Return a root struct that references a few entry points.
	root := &StructValue{
		Fields: make([]TypedValue, 10),
	}
	for i := range root.Fields {
		idx := (i * len(all)) / len(root.Fields)
		root.Fields[i] = TypedValue{T: IntType, V: all[idx]}
	}
	return root
}

func benchGCVisit(b *testing.B, nObjects int) {
	root := buildGCGraph(nObjects)
	alloc := NewAllocator(1 << 40) // very large limit
	var gcCycle int64
	var lastVisitCount int64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alloc.Reset()
		gcCycle++
		var visitCount int64
		vis := GCVisitorFn(gcCycle, alloc, &visitCount)
		vis(root)
		lastVisitCount = visitCount
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(lastVisitCount), "ns/visit")
	b.ReportMetric(float64(lastVisitCount), "visits/op")
}

func BenchmarkGCVisit_100(b *testing.B)     { benchGCVisit(b, 100) }
func BenchmarkGCVisit_1000(b *testing.B)    { benchGCVisit(b, 1000) }
func BenchmarkGCVisit_10000(b *testing.B)   { benchGCVisit(b, 10000) }
func BenchmarkGCVisit_100000(b *testing.B)  { benchGCVisit(b, 100000) }
func BenchmarkGCVisit_1000000(b *testing.B)  { benchGCVisit(b, 1000000) }
func BenchmarkGCVisit_10000000(b *testing.B) { benchGCVisit(b, 10000000) }
