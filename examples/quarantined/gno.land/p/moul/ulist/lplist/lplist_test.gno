package lplist

import (
	"testing"

	"gno.land/p/moul/ulist"
)

// TestLayeredProxyListBasicOperations tests the basic operations of LayeredProxyList
func TestLayeredProxyListBasicOperations(t *testing.T) {
	// Create source list with initial data
	source := ulist.New()
	source.Append(1, 2, 3)

	// Create proxy list with a simple multiplier migrator
	migrator := func(v any) any {
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
		want  any
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
	source := ulist.New()
	source.Append(1, 2, 3)
	proxy := NewLayeredProxyList(source, nil)
	proxy.Append(4, 5)

	// Test deleting from source (should fail)
	if err := proxy.Delete(1); err == nil {
		t.Error("Delete from source should return error")
	}

	// Test deleting from target (should succeed)
	if err := proxy.Delete(3); err != nil {
		t.Errorf("Delete from target failed: %s", err.Error())
	}

	// After deletion, the value might be undefined rather than nil
	// Check that it's not equal to the original value
	if got := proxy.Get(3); got == 5 {
		t.Errorf("Get(3) after delete = %v, want it to be deleted", got)
	}
}

// TestLayeredProxyListIteration tests iteration methods
func TestLayeredProxyListIteration(t *testing.T) {
	source := ulist.New()
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
	var values []any
	proxy.IteratorByOffset(1, 3, func(index int, value any) bool {
		values = append(values, value)
		return false
	})
	if len(values) != 3 {
		t.Errorf("IteratorByOffset returned %v values, want 3", len(values))
	}
}

// TestLayeredProxyListMustOperations tests must operations
func TestLayeredProxyListMustOperations(t *testing.T) {
	source := ulist.New()
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
	source := ulist.New()
	source.Append(1, 2)
	proxy := NewLayeredProxyList(source, nil)

	if got := proxy.Get(0); got != 1 {
		t.Errorf("Get(0) with nil migrator = %v, want 1", got)
	}
}

// TestLayeredProxyListEmpty tests operations on empty lists
func TestLayeredProxyListEmpty(t *testing.T) {
	source := ulist.New()
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

// TestLayeredProxyListChaining tests chaining of layered proxies with struct migrations
func TestLayeredProxyListChaining(t *testing.T) {
	// Define struct types for different versions
	type v1 struct {
		namev1 string
	}
	type v2 struct {
		namev2 string
	}
	type v3 struct {
		namev3 string
	}

	// Create source list with v1 objects
	source := ulist.New()
	source.Append(v1{namev1: "object1"}, v1{namev1: "object2"})

	// Migration function from v1 to v2
	v1Tov2 := func(v any) any {
		obj := v.(v1)
		return v2{namev2: obj.namev1 + "_v2"}
	}

	// Create first proxy with v1->v2 migration
	proxyV2 := NewLayeredProxyList(source, v1Tov2)
	proxyV2.Append(v2{namev2: "direct_v2"})

	// Migration function from v2 to v3
	v2Tov3 := func(v any) any {
		obj := v.(v2)
		return v3{namev3: obj.namev2 + "_v3"}
	}

	// Create second proxy with v2->v3 migration, using the first proxy as source
	proxyV3 := NewLayeredProxyList(proxyV2, v2Tov3)
	proxyV3.Append(v3{namev3: "direct_v3"})

	// Verify sizes
	if got := proxyV3.Size(); got != 4 {
		t.Errorf("proxyV3.Size() = %v, want 4", got)
	}

	// Test that all objects are correctly migrated when accessed through proxyV3
	expected := []struct {
		index int
		name  string
	}{
		{0, "object1_v2_v3"}, // v1 -> v2 -> v3
		{1, "object2_v2_v3"}, // v1 -> v2 -> v3
		{2, "direct_v2_v3"},  // v2 -> v3
		{3, "direct_v3"},     // v3 (no migration)
	}

	for _, tt := range expected {
		obj := proxyV3.Get(tt.index).(v3)
		if obj.namev3 != tt.name {
			t.Errorf("proxyV3.Get(%d).namev3 = %v, want %v", tt.index, obj.namev3, tt.name)
		}
	}

	// Verify getting items from middle layer (proxyV2)
	middleExpected := []struct {
		index int
		name  string
	}{
		{0, "object1_v2"}, // v1 -> v2
		{1, "object2_v2"}, // v1 -> v2
		{2, "direct_v2"},  // v2 (no migration)
	}

	for _, tt := range middleExpected {
		obj := proxyV2.Get(tt.index).(v2)
		if obj.namev2 != tt.name {
			t.Errorf("proxyV2.Get(%d).namev2 = %v, want %v", tt.index, obj.namev2, tt.name)
		}
	}
}
