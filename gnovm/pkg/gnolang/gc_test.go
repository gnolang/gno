package gnolang

import (
	"testing"
)

// Helper function to create a TypedValue from a TestValue
func newTestTypedValue() TypedValue {
	return TypedValue{V: PointerValue{TV: &TypedValue{V: &StructValue{
		Fields: nil,
	}}}}
}

func TestAddAndRemoveRoot(t *testing.T) {
	h := NewHeap()

	obj1 := NewObject(newTestTypedValue())
	h.AddRoot(obj1)

	if len(h.roots) != 1 {
		t.Errorf("Expected 1 root, got %d", len(h.roots))
	}

	h.RemoveRoot(obj1)

	if len(h.roots) != 0 {
		t.Errorf("Expected 0 roots, got %d", len(h.roots))
	}
}

func TestMarkAndSweep(t *testing.T) {
	h := NewHeap()

	// Create objects
	obj1 := NewObject(newTestTypedValue()) // root1
	obj2 := NewObject(newTestTypedValue()) // child1
	obj3 := NewObject(newTestTypedValue()) // child2
	obj4 := NewObject(newTestTypedValue()) // unreferenced

	// Add objects to heap
	h.AddObject(obj1)
	h.AddObject(obj2)
	h.AddObject(obj3)
	h.AddObject(obj4)

	// Set up references
	obj1.AddRef(obj2)
	obj2.AddRef(obj3)

	// Add root
	h.AddRoot(obj1)

	// Run GC
	deletedObjects := h.MarkAndSweep()

	if len(deletedObjects) != 1 {
		t.Errorf("Expected 1 deleted object, got %d", len(deletedObjects))
	}

	if ptr, ok := deletedObjects[0].tv.V.(PointerValue); ok && ptr == obj4.tv.V {
		t.Errorf("Expected 'unreferenced' to be deleted, but got '%s'", ptr)
	}
}

func TestCircularReference(t *testing.T) {
	h := NewHeap()

	// Create objects
	obj1 := NewObject(newTestTypedValue()) // root1
	obj2 := NewObject(newTestTypedValue()) // child1

	// Add objects to heap
	h.AddObject(obj1)
	h.AddObject(obj2)

	// Set up circular reference
	obj1.AddRef(obj2)
	obj2.AddRef(obj1)

	// Add root
	h.AddRoot(obj1)

	// Run GC
	deletedObjects := h.MarkAndSweep()

	if len(deletedObjects) != 0 {
		t.Errorf("Expected 0 deleted objects, got %d", len(deletedObjects))
	}
}

func TestDoubleFree(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on double free, but did not get one")
		}
	}()

	h := NewHeap()

	obj1 := NewObject(newTestTypedValue()) // root1
	h.AddObject(obj1)
	h.AddRoot(obj1)

	// Run GC to remove all references
	h.MarkAndSweep()

	// Remove root and try to sweep again
	h.RemoveRoot(obj1)

	// This should cause a panic due to "double free" as the object is already deleted
	h.MarkAndSweep()
}

func TestRootNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when removing a non-existent root, but did not get one")
		}
	}()

	h := NewHeap()

	obj1 := NewObject(newTestTypedValue()) // root1
	obj2 := NewObject(newTestTypedValue()) // root2

	h.AddObject(obj1)
	h.AddObject(obj2)
	h.AddRoot(obj1)

	// Attempt to remove a root that is not in the list
	h.RemoveRoot(obj2) // This should panic
}
