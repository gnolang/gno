package list

import (
	"testing"

	"gno.land/p/nt/ufmt/v0"
)

func TestList_Basic(t *testing.T) {
	var l List

	// Test empty list
	if l.Len() != 0 {
		t.Errorf("new list should be empty, got len %d", l.Len())
	}

	// Test append and length
	l.Append(1, 2, 3)
	if l.Len() != 3 {
		t.Errorf("expected len 3, got %d", l.Len())
	}

	// Test get
	if v := l.Get(0); v != 1 {
		t.Errorf("expected 1 at index 0, got %v", v)
	}
	if v := l.Get(1); v != 2 {
		t.Errorf("expected 2 at index 1, got %v", v)
	}
	if v := l.Get(2); v != 3 {
		t.Errorf("expected 3 at index 2, got %v", v)
	}

	// Test out of bounds
	if v := l.Get(-1); v != nil {
		t.Errorf("expected nil for negative index, got %v", v)
	}
	if v := l.Get(3); v != nil {
		t.Errorf("expected nil for out of bounds index, got %v", v)
	}
}

func TestList_Set(t *testing.T) {
	var l List
	l.Append(1, 2, 3)

	// Test valid set within bounds
	if ok := l.Set(1, 42); !ok {
		t.Error("Set should return true for valid index")
	}
	if v := l.Get(1); v != 42 {
		t.Errorf("expected 42 after Set, got %v", v)
	}

	// Test set at size (append)
	if ok := l.Set(3, 4); !ok {
		t.Error("Set should return true when appending at size")
	}
	if v := l.Get(3); v != 4 {
		t.Errorf("expected 4 after Set at size, got %v", v)
	}

	// Test invalid sets
	if ok := l.Set(-1, 10); ok {
		t.Error("Set should return false for negative index")
	}
	if ok := l.Set(5, 10); ok {
		t.Error("Set should return false for index > size")
	}

	// Verify list state hasn't changed after invalid operations
	expected := []any{1, 42, 3, 4}
	for i, want := range expected {
		if got := l.Get(i); got != want {
			t.Errorf("index %d = %v; want %v", i, got, want)
		}
	}
}

func TestList_Delete(t *testing.T) {
	var l List
	l.Append(1, 2, 3)

	// Test valid delete
	if v, ok := l.Delete(1); !ok || v != 2 {
		t.Errorf("Delete(1) = %v, %v; want 2, true", v, ok)
	}
	if l.Len() != 2 {
		t.Errorf("expected len 2 after delete, got %d", l.Len())
	}
	if v := l.Get(1); v != 3 {
		t.Errorf("expected 3 at index 1 after delete, got %v", v)
	}

	// Test invalid delete
	if v, ok := l.Delete(-1); ok || v != nil {
		t.Errorf("Delete(-1) = %v, %v; want nil, false", v, ok)
	}
	if v, ok := l.Delete(2); ok || v != nil {
		t.Errorf("Delete(2) = %v, %v; want nil, false", v, ok)
	}
}

func TestList_Slice(t *testing.T) {
	var l List
	l.Append(1, 2, 3, 4, 5)

	// Test valid ranges
	values := l.Slice(1, 4)
	expected := []any{2, 3, 4}
	if !sliceEqual(values, expected) {
		t.Errorf("Slice(1,4) = %v; want %v", values, expected)
	}

	// Test edge cases
	if values := l.Slice(-1, 2); !sliceEqual(values, []any{1, 2}) {
		t.Errorf("Slice(-1,2) = %v; want [1 2]", values)
	}
	if values := l.Slice(3, 10); !sliceEqual(values, []any{4, 5}) {
		t.Errorf("Slice(3,10) = %v; want [4 5]", values)
	}
	if values := l.Slice(3, 2); values != nil {
		t.Errorf("Slice(3,2) = %v; want nil", values)
	}
}

func TestList_ForEach(t *testing.T) {
	var l List
	l.Append(1, 2, 3)

	sum := 0
	l.ForEach(func(index int, value any) bool {
		sum += value.(int)
		return false
	})

	if sum != 6 {
		t.Errorf("ForEach sum = %d; want 6", sum)
	}

	// Test early termination
	count := 0
	l.ForEach(func(index int, value any) bool {
		count++
		return true // stop after first item
	})

	if count != 1 {
		t.Errorf("ForEach early termination count = %d; want 1", count)
	}
}

func TestList_Clone(t *testing.T) {
	var l List
	l.Append(1, 2, 3)

	clone := l.Clone()

	// Test same length
	if clone.Len() != l.Len() {
		t.Errorf("clone.Len() = %d; want %d", clone.Len(), l.Len())
	}

	// Test same values
	for i := 0; i < l.Len(); i++ {
		if clone.Get(i) != l.Get(i) {
			t.Errorf("clone.Get(%d) = %v; want %v", i, clone.Get(i), l.Get(i))
		}
	}

	// Test independence
	l.Set(0, 42)
	if clone.Get(0) == l.Get(0) {
		t.Error("clone should be independent of original")
	}
}

func TestList_DeleteRange(t *testing.T) {
	var l List
	l.Append(1, 2, 3, 4, 5)

	// Test valid range delete
	deleted := l.DeleteRange(1, 4)
	if deleted != 3 {
		t.Errorf("DeleteRange(1,4) deleted %d elements; want 3", deleted)
	}
	if l.Len() != 2 {
		t.Errorf("after DeleteRange(1,4) len = %d; want 2", l.Len())
	}
	expected := []any{1, 5}
	for i, want := range expected {
		if got := l.Get(i); got != want {
			t.Errorf("after DeleteRange(1,4) index %d = %v; want %v", i, got, want)
		}
	}

	// Test edge cases
	l = List{}
	l.Append(1, 2, 3)

	// Delete with negative start
	if deleted := l.DeleteRange(-1, 2); deleted != 2 {
		t.Errorf("DeleteRange(-1,2) deleted %d elements; want 2", deleted)
	}

	// Delete with end > length
	l = List{}
	l.Append(1, 2, 3)
	if deleted := l.DeleteRange(1, 5); deleted != 2 {
		t.Errorf("DeleteRange(1,5) deleted %d elements; want 2", deleted)
	}

	// Delete invalid range
	if deleted := l.DeleteRange(2, 1); deleted != 0 {
		t.Errorf("DeleteRange(2,1) deleted %d elements; want 0", deleted)
	}

	// Delete empty range
	if deleted := l.DeleteRange(1, 1); deleted != 0 {
		t.Errorf("DeleteRange(1,1) deleted %d elements; want 0", deleted)
	}
}

func TestList_Tree(t *testing.T) {
	var l List
	l.Append(1, 2, 3)

	rotree := l.Tree()
	expected := 3

	// Reverse iterate through the ReadOnlyTree
	rotree.ReverseIterateByOffset(0, rotree.Size(), func(key string, value any) bool {
		intValue := value.(int)
		if intValue != expected {
			t.Errorf("ReadOnlyTree expected %d, got %d", expected, intValue)
		}
		expected--
		return false
	})
}

func TestList_EmptyOperations(t *testing.T) {
	var l List

	// Operations on empty list
	if v := l.Get(0); v != nil {
		t.Errorf("Get(0) on empty list = %v; want nil", v)
	}

	// Set should work at index 0 for empty list (append case)
	if ok := l.Set(0, 1); !ok {
		t.Error("Set(0,1) on empty list = false; want true")
	}
	if v := l.Get(0); v != 1 {
		t.Errorf("Get(0) after Set = %v; want 1", v)
	}

	l = List{} // Reset to empty list
	if v, ok := l.Delete(0); ok || v != nil {
		t.Errorf("Delete(0) on empty list = %v, %v; want nil, false", v, ok)
	}
	if values := l.Slice(0, 1); values != nil {
		t.Errorf("Range(0,1) on empty list = %v; want nil", values)
	}
}

func TestList_DifferentTypes(t *testing.T) {
	var l List

	// Test with different types
	l.Append(42, "hello", true, 3.14)

	if v := l.Get(0).(int); v != 42 {
		t.Errorf("Get(0) = %v; want 42", v)
	}
	if v := l.Get(1).(string); v != "hello" {
		t.Errorf("Get(1) = %v; want 'hello'", v)
	}
	if v := l.Get(2).(bool); !v {
		t.Errorf("Get(2) = %v; want true", v)
	}
	if v := l.Get(3).(float64); v != 3.14 {
		t.Errorf("Get(3) = %v; want 3.14", v)
	}
}

func TestList_LargeOperations(t *testing.T) {
	var l List

	// Test with larger number of elements
	n := 1000
	for i := 0; i < n; i++ {
		l.Append(i)
	}

	if l.Len() != n {
		t.Errorf("Len() = %d; want %d", l.Len(), n)
	}

	// Test range on large list
	values := l.Slice(n-3, n)
	expected := []any{n - 3, n - 2, n - 1}
	if !sliceEqual(values, expected) {
		t.Errorf("Range(%d,%d) = %v; want %v", n-3, n, values, expected)
	}

	// Test large range deletion
	deleted := l.DeleteRange(100, 900)
	if deleted != 800 {
		t.Errorf("DeleteRange(100,900) = %d; want 800", deleted)
	}
	if l.Len() != 200 {
		t.Errorf("Len() after large delete = %d; want 200", l.Len())
	}
}

func TestList_ChainedOperations(t *testing.T) {
	var l List

	// Test sequence of operations
	l.Append(1, 2, 3)
	l.Delete(1)
	l.Append(4)
	l.Set(1, 5)

	expected := []any{1, 5, 4}
	for i, want := range expected {
		if got := l.Get(i); got != want {
			t.Errorf("index %d = %v; want %v", i, got, want)
		}
	}
}

func TestList_RangeEdgeCases(t *testing.T) {
	var l List
	l.Append(1, 2, 3, 4, 5)

	// Test various edge cases for Range
	cases := []struct {
		start, end int
		want       []any
	}{
		{-10, 2, []any{1, 2}},
		{3, 10, []any{4, 5}},
		{0, 0, nil},
		{5, 5, nil},
		{4, 3, nil},
		{-1, -1, nil},
	}

	for _, tc := range cases {
		got := l.Slice(tc.start, tc.end)
		if !sliceEqual(got, tc.want) {
			t.Errorf("Slice(%d,%d) = %v; want %v", tc.start, tc.end, got, tc.want)
		}
	}
}

func TestList_IndexConsistency(t *testing.T) {
	var l List

	// Initial additions
	l.Append(1, 2, 3, 4, 5) // [1,2,3,4,5]

	// Delete from middle
	l.Delete(2) // [1,2,4,5]

	// Add more elements
	l.Append(6, 7) // [1,2,4,5,6,7]

	// Delete range from middle
	l.DeleteRange(1, 4) // [1,6,7]

	// Add more elements
	l.Append(8, 9, 10) // [1,6,7,8,9,10]

	// Verify sequence is continuous
	expected := []any{1, 6, 7, 8, 9, 10}
	for i, want := range expected {
		if got := l.Get(i); got != want {
			t.Errorf("index %d = %v; want %v", i, got, want)
		}
	}

	// Verify no extra elements exist
	if l.Len() != len(expected) {
		t.Errorf("length = %d; want %d", l.Len(), len(expected))
	}

	// Verify all indices are accessible
	allValues := l.Slice(0, l.Len())
	if !sliceEqual(allValues, expected) {
		t.Errorf("Slice(0, Len()) = %v; want %v", allValues, expected)
	}

	// Verify no gaps in iteration
	var iteratedValues []any
	var indices []int
	l.ForEach(func(index int, value any) bool {
		iteratedValues = append(iteratedValues, value)
		indices = append(indices, index)
		return false
	})

	// Check values from iteration
	if !sliceEqual(iteratedValues, expected) {
		t.Errorf("ForEach values = %v; want %v", iteratedValues, expected)
	}

	// Check indices are sequential
	for i, idx := range indices {
		if idx != i {
			t.Errorf("ForEach index %d = %d; want %d", i, idx, i)
		}
	}
}

func TestList_RecursiveSafety(t *testing.T) {
	// Create a new list
	l := &List{}

	// Add some initial values
	l.Append("id1")
	l.Append("id2")
	l.Append("id3")

	// Test deep list traversal
	found := false
	l.ForEach(func(i int, v any) bool {
		if str, ok := v.(string); ok {
			if str == "id2" {
				found = true
				return true // stop iteration
			}
		}
		return false // continue iteration
	})

	if !found {
		t.Error("Failed to find expected value in list")
	}

	short := testing.Short()

	// Test recursive safety by performing multiple operations
	for i := 0; i < 1000; i++ {
		// Add new value
		l.Append(ufmt.Sprintf("id%d", i+4))

		if !short {
			// Search for a value
			var lastFound bool
			l.ForEach(func(j int, v any) bool {
				if str, ok := v.(string); ok {
					if str == ufmt.Sprintf("id%d", i+3) {
						lastFound = true
						return true
					}
				}
				return false
			})

			if !lastFound {
				t.Errorf("Failed to find value id%d after insertion", i+3)
			}
		}
	}

	// Verify final length
	expectedLen := 1003 // 3 initial + 1000 added
	if l.Len() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, l.Len())
	}

	if short {
		t.Skip("skipping extended recursive safety test in short mode")
	}
}

// Helper function to compare slices
func sliceEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
