package bitmap

import (
	"testing"
)

func TestBitmap_New(t *testing.T) {
	size := uint64(16)
	bm := New(size)

	if len(bm.data)*8 != int(size) {
		t.Errorf("Expected bitmap size: %d, got: %d", size, len(bm.data)*8)
	}
}

func TestBitmap_SetAndGet(t *testing.T) {
	bm := New(16)

	// Test setting and getting bits within bounds
	err := bm.Set(2)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !bm.Get(2) {
		t.Errorf("Expected bit at index 2 to be set")
	}

	err = bm.Set(15)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !bm.Get(15) {
		t.Errorf("Expected bit at index 15 to be set")
	}

	// Test setting and getting bits out of bounds
	err = bm.Set(16)
	if err == nil {
		t.Errorf("Expected error, got: nil")
	}
}

func TestBitmap_MustSet(t *testing.T) {
	bm := New(16)

	// Test setting bits within bounds
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got panic: %v", r)
		}
	}()

	bm.MustSet(2)
	if !bm.Get(2) {
		t.Errorf("Expected bit at index 2 to be set")
	}

	bm.MustSet(15)
	if !bm.Get(15) {
		t.Errorf("Expected bit at index 15 to be set")
	}

	// Test setting bits out of bounds
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but no panic occurred")
		}
	}()

	bm.MustSet(16 * 8)
}

func TestBitmap_Get(t *testing.T) {
	bm := New(16)

	// Test getting bits within bounds
	bm.MustSet(2)
	if !bm.Get(2) {
		t.Errorf("Expected bit at index 2 to be set")
	}

	if bm.Get(3) {
		t.Errorf("Expected bit at index 3 to be unset")
	}

	// Test getting bits out of bounds
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but no panic occurred")
		}
	}()

	bm.Get(16 * 8)
}
