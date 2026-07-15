package hashmap

import (
	"strconv"
	"testing"
)

func TestBasicOps(t *testing.T) {
	m := New()

	if m.Size() != 0 {
		t.Fatalf("new map size = %d, want 0", m.Size())
	}
	if m.Get("missing") != nil {
		t.Fatal("Get on empty map returned non-nil")
	}
	if m.Has("missing") {
		t.Fatal("Has on empty map returned true")
	}
	if _, removed := m.Remove("missing"); removed {
		t.Fatal("Remove on empty map reported removal")
	}

	if updated := m.Set("a", int64(1)); updated {
		t.Fatal("first Set reported update")
	}
	if updated := m.Set("a", int64(2)); !updated {
		t.Fatal("second Set did not report update")
	}
	if m.Size() != 1 {
		t.Fatalf("size = %d, want 1", m.Size())
	}
	if v := m.Get("a"); v == nil || v.(int64) != 2 {
		t.Fatalf("Get(a) = %v; want 2", v)
	}

	v, removed := m.Remove("a")
	if !removed || v.(int64) != 2 {
		t.Fatalf("Remove(a) = %v, %v; want 2, true", v, removed)
	}
	if m.Size() != 0 {
		t.Fatalf("size after remove = %d, want 0", m.Size())
	}
	if m.Has("a") {
		t.Fatal("Has(a) true after removal")
	}
}

func TestManyKeysAndCollisions(t *testing.T) {
	// 4 buckets + 100 keys forces many collisions per bucket.
	m := NewWithBuckets(4)
	for i := 0; i < 100; i++ {
		key := "key" + string(rune('A'+i%26)) + strconv.Itoa(i)
		if updated := m.Set(key, int64(i)); updated {
			t.Fatalf("unexpected update for fresh key %q", key)
		}
	}
	if m.Size() != 100 {
		t.Fatalf("size = %d, want 100", m.Size())
	}
	for i := 0; i < 100; i++ {
		key := "key" + string(rune('A'+i%26)) + strconv.Itoa(i)
		if v := m.Get(key); v == nil || v.(int64) != int64(i) {
			t.Fatalf("Get(%q) = %v; want %d", key, v, i)
		}
	}
	for i := 0; i < 100; i += 2 {
		key := "key" + string(rune('A'+i%26)) + strconv.Itoa(i)
		if _, removed := m.Remove(key); !removed {
			t.Fatalf("Remove(%q) missed", key)
		}
	}
	if m.Size() != 50 {
		t.Fatalf("size after removals = %d, want 50", m.Size())
	}
}

func TestIterate(t *testing.T) {
	m := NewWithBuckets(8)
	want := map[string]int64{"x": 1, "y": 2, "z": 3}
	for k, v := range want {
		m.Set(k, v)
	}

	seen := map[string]int64{}
	stopped := m.Iterate(func(k string, v any) bool {
		seen[k] = v.(int64)
		return false
	})
	if stopped {
		t.Fatal("full iteration reported early stop")
	}
	if len(seen) != len(want) {
		t.Fatalf("iterated %d entries, want %d", len(seen), len(want))
	}
	for k, v := range want {
		if seen[k] != v {
			t.Fatalf("seen[%q] = %d, want %d", k, seen[k], v)
		}
	}

	// Early stop after the first entry.
	n := 0
	stopped = m.Iterate(func(k string, v any) bool {
		n++
		return true
	})
	if !stopped || n != 1 {
		t.Fatalf("early stop: stopped=%v n=%d; want true, 1", stopped, n)
	}
}

func TestIterateDeterministic(t *testing.T) {
	build := func() []string {
		m := NewWithBuckets(16)
		for i := 0; i < 50; i++ {
			m.Set("k"+strconv.Itoa(i), i)
		}
		var order []string
		m.Iterate(func(k string, v any) bool {
			order = append(order, k)
			return false
		})
		return order
	}
	a, b := build(), build()
	if len(a) != len(b) || len(a) != 50 {
		t.Fatalf("orders have %d and %d entries, want 50", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("iteration order diverged at %d: %q vs %q", i, a[i], b[i])
		}
	}
}

func TestBucketCountValidation(t *testing.T) {
	for _, bad := range []int{0, -1, 3, 12, 1000} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewWithBuckets(%d) did not panic", bad)
				}
			}()
			NewWithBuckets(bad)
		}()
	}
	for _, good := range []int{1, 2, 64, 1024} {
		if m := NewWithBuckets(good); m == nil {
			t.Fatalf("NewWithBuckets(%d) returned nil", good)
		}
	}
}
