package gnolang

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

// TestStringRangeSet_Model checks the treap against a brute-force sorted
// slice model under a fixed-seed random workload of inserts, containment
// and overlap queries, removals and retain() cycles.
func TestStringRangeSet_Model(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	var set stringRangeSet
	var model []stringRange // sorted by start, disjoint

	modelContaining := func(p uintptr) *stringRange {
		for i := range model {
			if p >= model[i].start && p < model[i].end {
				return &model[i]
			}
		}
		return nil
	}
	modelOverlapping := func(p, end uintptr) bool {
		for _, r := range model {
			if r.start < end && r.end > p {
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

	const space = 4096
	for step := 0; step < 20000; step++ {
		switch op := rng.Intn(10); {
		case op < 5: // insert a random non-overlapping range
			start := uintptr(rng.Intn(space))
			end := start + uintptr(1+rng.Intn(16))
			if modelOverlapping(start, end) {
				continue
			}
			r := stringRange{start: start, end: end}
			set.insert(r)
			model = append(model, r)
			sort.Slice(model, func(i, j int) bool { return model[i].start < model[j].start })
		case op < 7: // containment + overlap queries
			p := uintptr(rng.Intn(space + 32))
			want := modelContaining(p)
			got := set.containing(p)
			if (want == nil) != (got == nil) || (want != nil && *want != *got) {
				t.Fatalf("step %d: containing(%d): set=%v model=%v", step, p, got, want)
			}
			end := p + uintptr(1+rng.Intn(16))
			if got, want := set.overlapping(p, end) != nil, modelOverlapping(p, end); got != want {
				t.Fatalf("step %d: overlapping(%d,%d): set=%v model=%v", step, p, end, got, want)
			}
		case op < 8: // remove a random existing range
			if len(model) == 0 {
				continue
			}
			i := rng.Intn(len(model))
			set.remove(model[i].start)
			model = append(model[:i], model[i+1:]...)
		case op < 9: // stamp a random subset with cycle, then retain it
			cycle := int64(step + 1)
			for i := range model {
				if rng.Intn(2) == 0 {
					model[i].lastCycle = cycle
					r := set.containing(model[i].start)
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
		default: // remove a non-existent key is a no-op
			set.remove(uintptr(space + 100 + rng.Intn(100)))
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
			for i := 0; i < live; i++ {
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
			for i := 0; i < live; i++ {
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
