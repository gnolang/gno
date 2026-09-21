package gnolang

import "testing"

// benchGCStackScan calibrates gcStackScanSlopeGas: the per-slot cost of
// GarbageCollect's linear walks over m.Values and Allocator.anchors when the
// slots are primitives (V == nil), which produce no visit and so are not
// covered by gcVisitGas. Mirrors benchGCVisit's shape and reporting.
func benchGCStackScan(b *testing.B, nSlots int) {
	b.Helper()
	values := make([]TypedValue, nSlots)
	for i := range values {
		values[i] = TypedValue{T: IntType} // V == nil
	}
	alloc := NewAllocator(1 << 40) // very large limit
	var gcCycle int64

	b.ResetTimer()
	for range b.N {
		alloc.Reset()
		gcCycle++
		var visitCount int64
		vis := GCVisitorFn(gcCycle, alloc, &visitCount)
		// Exactly the walk GarbageCollect runs over m.Values.
		for j := range values {
			alloc.Recount(allocTypedValue)
			if v := values[j].V; v != nil {
				if stop := vis(v); stop {
					break
				}
			}
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(nSlots), "ns/slot")
	b.ReportMetric(float64(nSlots), "slots/op")
}

func BenchmarkGCStackScan_1000(b *testing.B)     { benchGCStackScan(b, 1000) }
func BenchmarkGCStackScan_10000(b *testing.B)    { benchGCStackScan(b, 10000) }
func BenchmarkGCStackScan_100000(b *testing.B)   { benchGCStackScan(b, 100000) }
func BenchmarkGCStackScan_1000000(b *testing.B)  { benchGCStackScan(b, 1000000) }
func BenchmarkGCStackScan_10000000(b *testing.B) { benchGCStackScan(b, 10000000) }
