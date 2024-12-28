package ulist

import (
	"testing"
)

// TestLayeredProxyListBasicOperations tests the basic operations of LayeredProxyList
func TestLayeredProxyListBasicOperations(t *testing.T) {
	// Create source list with initial data
	source := New()
	source.Append(1, 2, 3)

	// Create proxy list with a simple multiplier migrator
	migrator := func(v interface{}) interface{} {
		return v.(int) * 2
	}
	proxy := NewLayeredProxyList(source, migrator)

	// Test initial state
	if got := proxy.Size(); got != 3 {
		t.Errorf("initial Size() = %v, want %v", got, 3)
	}
	if got := proxy.TotalSize(); got != 3 {
		t.Errorf("initial TotalSize() = %v, want %v", got, 3)
	}

	// Test Get with migration
	tests := []struct {
		index int
		want  interface{}
	}{
		{0, 2}, // 1 * 2
		{1, 4}, // 2 * 2
		{2, 6}, // 3 * 2
	}

	for _, tt := range tests {
		if got := proxy.Get(tt.index); got != tt.want {
			t.Errorf("Get(%v) = %v, want %v", tt.index, got, tt.want)
		}
	}

	// Test Append to target
	proxy.Append(7, 8)
	if got := proxy.Size(); got != 5 {
		t.Errorf("Size() after append = %v, want %v", got, 5)
	}

	// Test Get from target (no migration)
	if got := proxy.Get(3); got != 7 {
		t.Errorf("Get(3) = %v, want %v", got, 7)
	}
}

// TestLayeredProxyListDelete tests delete operations
func TestLayeredProxyListDelete(t *testing.T) {
	source := New()
	source.Append(1, 2, 3)
	proxy := NewLayeredProxyList(source, nil)
	proxy.Append(4, 5)

	// Test delete from source (should fail)
	if err := proxy.Delete(1); err == nil {
		t.Error("Delete from source should return error")
	}

	// Test delete from target (should succeed)
	if err := proxy.Delete(3); err != nil {
		t.Errorf("Delete from target failed: %v", err)
	}

	// Verify deletion
	if got := proxy.Get(3); got != nil {
		t.Errorf("Get(3) after delete = %v, want nil", got)
	}
}

// TestLayeredProxyListIteration tests iteration methods
func TestLayeredProxyListIteration(t *testing.T) {
	source := New()
	source.Append(1, 2, 3)
	proxy := NewLayeredProxyList(source, nil)
	proxy.Append(4, 5)

	// Test GetRange
	entries := proxy.GetRange(0, 4)
	if len(entries) != 5 {
		t.Errorf("GetRange returned %v entries, want 5", len(entries))
	}

	// Test reverse iteration
	entries = proxy.GetRange(4, 0)
	if len(entries) != 5 {
		t.Errorf("Reverse GetRange returned %v entries, want 5", len(entries))
	}

	// Test IteratorByOffset with positive count
	var values []interface{}
	proxy.IteratorByOffset(1, 3, func(index int, value interface{}) bool {
		values = append(values, value)
		return false
	})
	if len(values) != 3 {
		t.Errorf("IteratorByOffset returned %v values, want 3", len(values))
	}
}

// TestLayeredProxyListMustOperations tests must operations
func TestLayeredProxyListMustOperations(t *testing.T) {
	source := New()
	source.Append(1, 2)
	proxy := NewLayeredProxyList(source, nil)

	// Test MustGet success
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustGet panicked unexpectedly: %v", r)
		}
	}()
	if got := proxy.MustGet(1); got != 2 {
		t.Errorf("MustGet(1) = %v, want 2", got)
	}

	// Test MustGet panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet should have panicked")
		}
	}()
	proxy.MustGet(99) // Should panic
}

// TestLayeredProxyListWithNilMigrator tests behavior without a migrator
func TestLayeredProxyListWithNilMigrator(t *testing.T) {
	source := New()
	source.Append(1, 2)
	proxy := NewLayeredProxyList(source, nil)

	if got := proxy.Get(0); got != 1 {
		t.Errorf("Get(0) with nil migrator = %v, want 1", got)
	}
}

// TestLayeredProxyListEmpty tests operations on empty lists
func TestLayeredProxyListEmpty(t *testing.T) {
	source := New()
	proxy := NewLayeredProxyList(source, nil)

	if got := proxy.Size(); got != 0 {
		t.Errorf("Size() of empty list = %v, want 0", got)
	}

	if got := proxy.Get(0); got != nil {
		t.Errorf("Get(0) of empty list = %v, want nil", got)
	}

	entries := proxy.GetRange(0, 10)
	if len(entries) != 0 {
		t.Errorf("GetRange on empty list returned %v entries, want 0", len(entries))
	}
}
