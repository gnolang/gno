package gnolang

import (
	"testing"
)

// Helper function to create a TypedValue from a TestValue
func newTestTypedValue() TypedValue {
	return Unwrap(TypedValue{V: PointerValue{TV: &TypedValue{V: &StructValue{
		Fields: nil,
	}}}})
}

func TestAddAndRemoveRoot(t *testing.T) {
	h := NewHeap()

	root := NewObject(newTestTypedValue())
	obj1 := NewObject(newTestTypedValue())
	visited := make(map[*GcObj]bool)
	h.AddRef(root, obj1, visited)

	h.AddRoot(root)

	if len(h.roots) != 1 {
		t.Errorf("Expected 1 root, got %d", len(h.roots))
	}

	h.RemoveRoot(obj1.tv)

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
	h.AddObject(obj2)
	h.AddObject(obj3)
	h.AddObject(obj4)

	// Set up references
	visited := make(map[*GcObj]bool)
	h.AddRef(obj1, obj2, visited)
	visited = make(map[*GcObj]bool)
	h.AddRef(obj1, obj3, visited)

	// Add root
	h.AddRoot(obj1)

	// Run GC
	deletedObjects := h.MarkAndSweep()

	if len(deletedObjects) != 1 {
		t.Errorf("Expected 1 deleted object, got %d", len(deletedObjects))
	}

	if strct, ok := deletedObjects[0].tv.V.(*StructValue); !ok || strct != obj4.tv.V {
		t.Errorf("Expected 'unreferenced' to be deleted, but got '%s'", strct)
	}
}

func TestAddRef(t *testing.T) {
	h := NewHeap()

	t.Run("StructValue", func(t *testing.T) {
		root := NewObject(TypedValue{T: nil, V: &StructValue{
			Fields: []TypedValue{
				{T: nil, V: StringValue("field1")},
				{T: nil, V: StringValue("field2")},
			},
		}})
		child := NewObject(newTestTypedValue())
		visited := make(map[*GcObj]bool)
		h.AddRef(root, child, visited)

		if len(root.refs) != 1 {
			t.Errorf("Expected 1 ref for StructValue, got %d", len(root.refs))
		}
		if root.refs[0] != child {
			t.Errorf("Expected child to be added as a reference")
		}
	})

	t.Run("SliceValue", func(t *testing.T) {
		root := NewObject(TypedValue{T: nil, V: &SliceValue{
			Base: &ArrayValue{
				List: []TypedValue{
					{T: nil, V: StringValue("test1")},
					{T: nil, V: StringValue("test2")},
				},
			},
		}})
		visited := make(map[*GcObj]bool)
		h.AddRef(root, root, visited)

		if len(root.refs) != 1 {
			t.Errorf("Expected 1 ref for SliceValue, got %d", len(root.refs))
		}
		if root.refs[0] != root {
			t.Errorf("Expected root to be added as a self-reference")
		}
	})

	t.Run("StringValue", func(t *testing.T) {
		root := NewObject(newTestTypedValue())
		stringObj := NewObject(TypedValue{T: nil, V: StringValue("test")})
		visited := make(map[*GcObj]bool)
		h.AddRef(root, stringObj, visited)

		if len(root.refs) != 1 {
			t.Errorf("Expected 1 ref for StringValue, got %d", len(root.refs))
		}
		if root.refs[0] != stringObj {
			t.Errorf("Expected stringObj to be added as a reference")
		}
	})

	t.Run("Default case (panic)", func(t *testing.T) {
		root := NewObject(TypedValue{T: nil, V: BigintValue{}})
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for unhandled type, but no panic occurred")
			}
		}()
		visited := make(map[*GcObj]bool)
		h.AddRef(root, root, visited)
	})
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
	visited := make(map[*GcObj]bool)
	h.AddRef(obj1, obj2, visited)
	visited = make(map[*GcObj]bool)
	h.AddRef(obj2, obj1, visited)

	// Add root
	h.AddRoot(obj1)

	// Run GC
	deletedObjects := h.MarkAndSweep()

	if len(deletedObjects) != 0 {
		t.Errorf("Expected 0 deleted objects, got %d", len(deletedObjects))
	}
}

func TestRootNotFound(t *testing.T) {
	h := NewHeap()

	root1 := NewObject(newTestTypedValue())
	root2 := NewObject(newTestTypedValue())

	h.AddObject(root1)
	h.AddObject(root2)
	h.AddRoot(root1)

	h.RemoveRoot(root2.tv)
}
