package gnolang

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

// TestStringRangeSet_Model checks the treap against a brute-force sorted
// slice model under a fixed-seed random workload of inserts, containment
// and overlap queries, and retain() cycles. Ranges are backed by real
// strings (extents derive from stringRange.str), so disjointness of live
// entries is guaranteed by the Go allocator itself; probe points are
// derived from live extents to exercise hits, boundary misses, and gaps.
func TestStringRangeSet_Model(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	var set stringRangeSet
	var model []stringRange // sorted by start, disjoint

	modelContaining := func(p uintptr) *stringRange {
		for i := range model {
			if s, e := model[i].extent(); p >= s && p < e {
				return &model[i]
			}
		}
		return nil
	}
	modelOverlapping := func(p, end uintptr) bool {
		for i := range model {
			if s, e := model[i].extent(); s < end && e > p {
				return true
			}
		}
		return false
	}
	check := func(step int) {
		got := set.ranges()
		if len(got) != len(model) || set.len() != len(model) {
			t.Fatalf("step %d: len mismatch: set=%d/%d model=%d", step, len(got), set.len(), len(model))
		}
		for i := range got {
			if got[i] != model[i] {
				t.Fatalf("step %d: range %d mismatch: set=%+v model=%+v", step, i, got[i], model[i])
			}
		}
	}
	// probe returns an address to query: usually anchored on a live
	// extent (start-1, inside, end-1, end, just past), sometimes far off.
	probe := func() uintptr {
		if len(model) == 0 || rng.Intn(8) == 0 {
			return uintptr(rng.Uint64()) // arbitrary address, almost surely a miss
		}
		s, e := model[rng.Intn(len(model))].extent()
		return s + uintptr(rng.Intn(int(e-s)+8)) - 4
	}

	for step := range 20000 {
		switch op := rng.Intn(10); {
		case op < 5: // insert a fresh string-backed range
			// len >= 2: Go interns 1-byte strings into a shared static
			// array, so two equal 1-byte strings would produce identical
			// (overlapping) extents. Production handles that shape via
			// trackString's clone-on-overlap; insert's contract assumes
			// disjoint, so the generator must not produce it.
			b := make([]byte, 2+rng.Intn(15))
			for i := range b {
				b[i] = byte('a' + rng.Intn(26))
			}
			r := stringRange{str: string(b)}
			set.insert(r)
			model = append(model, r)
			sort.Slice(model, func(i, j int) bool { return model[i].start() < model[j].start() })
		case op < 9: // containment + overlap queries
			p := probe()
			want := modelContaining(p)
			got := set.containing(p)
			if (want == nil) != (got == nil) || (want != nil && *want != *got) {
				t.Fatalf("step %d: containing(%d): set=%v model=%v", step, p, got, want)
			}
			end := p + uintptr(1+rng.Intn(16))
			if got, want := set.overlapping(p, end) != nil, modelOverlapping(p, end); got != want {
				t.Fatalf("step %d: overlapping(%d,%d): set=%v model=%v", step, p, end, got, want)
			}
		default: // stamp a random subset with cycle, then retain it
			cycle := int64(step + 1)
			for i := range model {
				if rng.Intn(2) == 0 {
					model[i].lastCycle = cycle
					r := set.containing(model[i].start())
					if r == nil {
						t.Fatalf("step %d: lost range %+v", step, model[i])
					}
					r.lastCycle = cycle
				}
			}
			set.retain(cycle)
			kept := model[:0]
			for _, r := range model {
				if r.lastCycle == cycle {
					kept = append(kept, r)
				}
			}
			model = kept
		}
		check(step)
	}
}

// BenchmarkNewStringTracked measures NewString as a function of the number
// of live tracked strings. With the treap this should stay roughly flat
// (logarithmic); a sorted-slice implementation grew linearly (memmove).
func BenchmarkNewStringTracked(b *testing.B) {
	for _, live := range []int{0, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("live=%d", live), func(b *testing.B) {
			alloc := NewAllocator(1 << 40)
			keep := make([]StringValue, 0, live)
			for i := range live {
				keep = append(keep, alloc.NewString(fmt.Sprintf("s%08d", i)))
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = alloc.NewString(fmt.Sprintf("x%08d", i))
			}
			_ = keep
		})
	}
}

// BenchmarkGCStringPass measures the GC-side cost (CountStringBytes over
// every live string, then CleanupTrackedStrings) per cycle.
func BenchmarkGCStringPass(b *testing.B) {
	for _, live := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("live=%d", live), func(b *testing.B) {
			alloc := NewAllocator(1 << 40)
			keep := make([]StringValue, 0, live)
			for i := range live {
				keep = append(keep, alloc.NewString(fmt.Sprintf("s%08d", i)))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cycle := int64(i + 1)
				for _, s := range keep {
					alloc.CountStringBytes(string(s), cycle)
				}
				alloc.CleanupTrackedStrings(cycle)
			}
		})
	}
}
