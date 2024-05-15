package bitmap

import (
	"testing"
)

func TestNewBitmap(t *testing.T) {
	// Test creating a bitmap with zero size
	bm := New(0)
	if len(bm.data) != 0 {
		t.Errorf("Expected empty data slice for size 0, got %d bytes", len(bm.data))
	}

	// Test creating a bitmap with a small size
	bm = New(8)
	if len(bm.data) != 1 {
		t.Errorf("Expected 1 byte for size 8, got %d bytes", len(bm.data))
	}

	// Test creating a bitmap with a size not divisible by 8
	bm = New(10)
	if len(bm.data) != 2 {
		t.Errorf("Expected 2 bytes for size 10, got %d bytes", len(bm.data))
	}
}

func TestSetGet(t *testing.T) {
	bm := New(16)

	// Test setting and getting a bit
	bm.Set(5)
	if !bm.Get(5) {
		t.Errorf("Expected bit 5 to be set after setting")
	}
	if bm.Get(10) {
		t.Errorf("Expected bit 10 to be unset")
	}

	// Test setting and getting out of bounds
	defer func() {
		if recover() == nil {
			t.Errorf("Expected panic for out of bounds index")
		}
	}()
	bm.Set(20) // Should panic

	defer func() {
		if recover() == nil {
			t.Errorf("Expected panic for out of bounds index")
		}
	}()
	bm.Get(20) // Should panic
}

func TestLargeBitmap(t *testing.T) {
	// Test creating and setting a bit in a larger bitmap
	bm := New(1024)
	bm.Set(512)
	if !bm.Get(512) {
		t.Errorf("Expected bit 512 to be set in a large bitmap")
	}
}
