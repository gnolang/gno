package deque

import "testing"

func TestNew(t *testing.T) {
	d := New()
	if d.Size() != 0 {
		t.Errorf("Expected size 0, got %d", d.Size())
	}
	if d.MaxSize() != 0 {
		t.Errorf("Expected max size 0 (unlimited), got %d", d.MaxSize())
	}
	if d.IsBounded() {
		t.Error("Expected unbounded deque")
	}
}

func TestNewBounded(t *testing.T) {
	d := NewBounded(5)
	if d.MaxSize() != 5 {
		t.Errorf("Expected max size 5, got %d", d.MaxSize())
	}
	if d.Size() != 0 {
		t.Errorf("Expected size 0, got %d", d.Size())
	}
	if !d.IsBounded() {
		t.Error("Expected bounded deque")
	}
}

func TestPushBackPopFront(t *testing.T) {
	d := New() // unbounded

	// Test empty pop
	if d.PopFront() != nil {
		t.Error("Expected nil when popping from empty deque")
	}

	d.PushBack("a", "b", "c")
	if d.Size() != 3 {
		t.Errorf("Expected size 3, got %d", d.Size())
	}

	item := d.PopFront()
	if item != "a" {
		t.Errorf("Expected 'a', got %v", item)
	}
	if d.Size() != 2 {
		t.Errorf("Expected size 2, got %d", d.Size())
	}
}

func TestPushFrontPopBack(t *testing.T) {
	d := New() // unbounded

	d.PushFront("a", "b", "c")
	if d.Size() != 3 {
		t.Errorf("Expected size 3, got %d", d.Size())
	}

	// Elements should be in reverse order: [c, b, a]
	if d.First() != "c" {
		t.Errorf("Expected 'c' as first, got %v", d.First())
	}
	if d.Last() != "a" {
		t.Errorf("Expected 'a' as last, got %v", d.Last())
	}

	item := d.PopBack()
	if item != "a" {
		t.Errorf("Expected 'a', got %v", item)
	}
	if d.Size() != 2 {
		t.Errorf("Expected size 2, got %d", d.Size())
	}
}

func TestBoundedEvictionBack(t *testing.T) {
	d := NewBounded(3)

	// Fill to capacity
	d.PushBack("a", "b", "c")
	if d.Size() != 3 {
		t.Errorf("Expected size 3, got %d", d.Size())
	}

	// Push beyond capacity - should evict "a" from front
	d.PushBack("d")
	if d.Size() != 3 {
		t.Errorf("Expected size 3, got %d", d.Size())
	}
	if d.First() != "b" {
		t.Errorf("Expected 'b' (oldest after eviction), got %v", d.First())
	}
	if d.Last() != "d" {
		t.Errorf("Expected 'd' (newest), got %v", d.Last())
	}
}

func TestBoundedEvictionFront(t *testing.T) {
	d := NewBounded(3)

	// Fill to capacity
	d.PushBack("a", "b", "c")

	// Push to front beyond capacity - should evict "c" from back
	d.PushFront("x")
	if d.Size() != 3 {
		t.Errorf("Expected size 3, got %d", d.Size())
	}
	if d.First() != "x" {
		t.Errorf("Expected 'x' (newest at front), got %v", d.First())
	}
	if d.Last() != "b" {
		t.Errorf("Expected 'b' (oldest after back eviction), got %v", d.Last())
	}
}

func TestFirstLast(t *testing.T) {
	d := New()

	// Empty deque
	if d.First() != nil {
		t.Error("Expected nil for empty deque first")
	}
	if d.Last() != nil {
		t.Error("Expected nil for empty deque last")
	}

	d.PushBack("x", "y", "z")
	if d.First() != "x" {
		t.Errorf("Expected 'x', got %v", d.First())
	}
	if d.Last() != "z" {
		t.Errorf("Expected 'z', got %v", d.Last())
	}
}

func TestGet(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c")

	if d.Get(0) != "a" {
		t.Errorf("Expected 'a' at index 0, got %v", d.Get(0))
	}
	if d.Get(1) != "b" {
		t.Errorf("Expected 'b' at index 1, got %v", d.Get(1))
	}
	if d.Get(2) != "c" {
		t.Errorf("Expected 'c' at index 2, got %v", d.Get(2))
	}
	if d.Get(3) != nil {
		t.Error("Expected nil for out-of-bounds index")
	}
	if d.Get(-1) != nil {
		t.Error("Expected nil for negative index")
	}
}

func TestList(t *testing.T) {
	d := New()

	// Empty deque
	list := d.List()
	if list != nil {
		t.Errorf("Expected nil for empty deque, got %v", list)
	}

	d.PushBack("x", "y", "z")
	list = d.List()
	if len(list) != 3 {
		t.Errorf("Expected length 3, got %d", len(list))
	}
	if list[0] != "x" || list[1] != "y" || list[2] != "z" {
		t.Errorf("Expected [x, y, z], got %v", list)
	}
}

func TestEnumerate(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c")

	var enumerated []string
	var indices []int

	d.Enumerate(func(index int, item any) bool {
		indices = append(indices, index)
		enumerated = append(enumerated, item.(string))
		return false // continue
	})

	if len(enumerated) != 3 {
		t.Errorf("Expected 3 enumerated items, got %d", len(enumerated))
	}
	if enumerated[0] != "a" || enumerated[1] != "b" || enumerated[2] != "c" {
		t.Errorf("Expected [a, b, c], got %v", enumerated)
	}
	if indices[0] != 0 || indices[1] != 1 || indices[2] != 2 {
		t.Errorf("Expected [0, 1, 2], got %v", indices)
	}
}

func TestEnumerateEarlyStop(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c")

	var enumerated []string

	d.Enumerate(func(index int, item any) bool {
		enumerated = append(enumerated, item.(string))
		return item == "b" // stop after b
	})

	if len(enumerated) != 2 {
		t.Errorf("Expected 2 enumerated items, got %d", len(enumerated))
	}
	if enumerated[0] != "a" || enumerated[1] != "b" {
		t.Errorf("Expected [a, b], got %v", enumerated)
	}
}

func TestSetMaxSize(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c", "d", "e")

	// Reduce max size - should trim from front
	d.SetMaxSize(3)
	if d.Size() != 3 {
		t.Errorf("Expected size 3 after reducing max size, got %d", d.Size())
	}
	if d.First() != "c" {
		t.Errorf("Expected 'c' as first after trim, got %v", d.First())
	}

	// Increase max size
	d.SetMaxSize(10)
	if d.MaxSize() != 10 {
		t.Errorf("Expected max size 10, got %d", d.MaxSize())
	}
}

func TestClear(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c")

	d.Clear()

	if d.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", d.Size())
	}
	if d.First() != nil {
		t.Error("Expected nil first after clear")
	}
	if d.Last() != nil {
		t.Error("Expected nil last after clear")
	}
	if !d.IsEmpty() {
		t.Error("Expected deque to be empty after clear")
	}
}

func TestUnboundedGrowth(t *testing.T) {
	d := New() // unbounded

	// Add many items
	for i := 0; i < 100; i++ {
		d.PushBack(i)
	}

	if d.Size() != 100 {
		t.Errorf("Expected size 100, got %d", d.Size())
	}
	if d.First() != 0 {
		t.Errorf("Expected 0 as first, got %v", d.First())
	}
	if d.Last() != 99 {
		t.Errorf("Expected 99 as last, got %v", d.Last())
	}
}

func TestMixedOperations(t *testing.T) {
	d := NewBounded(4)

	// Test mixed push front and back operations
	d.PushBack("b", "c")
	d.PushFront("a")
	d.PushBack("d")

	// Should have [a, b, c, d]
	list := d.List()
	if len(list) != 4 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" {
		t.Errorf("Expected [a, b, c, d], got %v", list)
	}

	// Push beyond capacity
	d.PushFront("x") // Should evict "d" from back

	// Should have [x, a, b, c]
	list = d.List()
	if len(list) != 4 || list[0] != "x" || list[1] != "a" || list[2] != "b" || list[3] != "c" {
		t.Errorf("Expected [x, a, b, c], got %v", list)
	}

	// Mixed pop operations
	front := d.PopFront() // "x"
	back := d.PopBack()   // "c"

	if front != "x" {
		t.Errorf("Expected 'x' from front, got %v", front)
	}
	if back != "c" {
		t.Errorf("Expected 'c' from back, got %v", back)
	}

	// Should have [a, b]
	list = d.List()
	if len(list) != 2 || list[0] != "a" || list[1] != "b" {
		t.Errorf("Expected [a, b], got %v", list)
	}
}

func TestAllIterator(t *testing.T) {
	d := New()
	d.PushBack("x", "y", "z")

	var items []any
	iter := d.All()
	iter(func(value any) bool {
		items = append(items, value)
		return true // continue
	})

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
	if items[0] != "x" || items[1] != "y" || items[2] != "z" {
		t.Errorf("Expected [x, y, z], got %v", items)
	}
}

func TestBackwardIterator(t *testing.T) {
	d := New()
	d.PushBack("x", "y", "z")

	var items []any
	iter := d.Backward()
	iter(func(value any) bool {
		items = append(items, value)
		return true // continue
	})

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
	if items[0] != "z" || items[1] != "y" || items[2] != "x" {
		t.Errorf("Expected [z, y, x] (reverse), got %v", items)
	}
}

func TestAllIteratorEarlyBreak(t *testing.T) {
	d := New()
	d.PushBack("a", "b", "c", "d", "e")

	var items []any
	iter := d.All()
	iter(func(value any) bool {
		items = append(items, value)
		if value == "c" {
			return false // stop
		}
		return true // continue
	})

	if len(items) != 3 {
		t.Errorf("Expected 3 items after break, got %d", len(items))
	}
	if items[0] != "a" || items[1] != "b" || items[2] != "c" {
		t.Errorf("Expected [a, b, c], got %v", items)
	}
}
