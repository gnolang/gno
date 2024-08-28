package gnolang

import (
	"fmt"
)

// GcObj represents an object in the heap.
type GcObj struct {
	marked bool
	tv     TypedValue
	refs   []*GcObj
}

func (obj *GcObj) AddRef(ref *GcObj) {
	obj.refs = append(obj.refs, ref)
}

// NewObject creates a new object with a given name.
func NewObject(tv TypedValue) *GcObj {
	return &GcObj{
		tv: tv,
	}
}

type Heap struct {
	objects []*GcObj
	roots   []*GcObj
}

func NewHeap() *Heap {
	return &Heap{}
}

func (h *Heap) FindObjectByTV(tv TypedValue) *GcObj {
	for _, object := range h.objects {
		if object.tv == tv {
			return object
		}
	}
	return nil
}

func (h *Heap) RemoveRoot(root *GcObj) {
	roots := make([]*GcObj, 0, len(h.roots))
	var deleted bool

	for _, r := range h.roots {
		if !deleted && r.tv == root.tv {
			fmt.Printf("REMOVEROOT: %+v\n", root.tv)
			deleted = true
			continue
		}
		roots = append(roots, root)
	}

	h.roots = roots
}

func (h *Heap) AddObject(obj *GcObj) {
	h.objects = append(h.objects, obj)
}

func (h *Heap) AddRoot(obj *GcObj) {
	h.roots = append(h.roots, obj)
}

func (h *Heap) MarkAndSweep() []*GcObj {
	// Mark phase: mark all reachable objects
	for _, root := range h.roots {
		h.mark(root)
	}

	// Sweep phase: remove unmarked objects
	return h.sweep()
}

// mark recursively marks all reachable objects starting from a root.
func (h *Heap) mark(obj *GcObj) {
	if obj == nil {
		return
	}
	if obj.marked {
		return
	}
	obj.marked = true
	fmt.Printf("Marking object: %s\n", obj.tv)

	for _, ref := range obj.refs {
		h.mark(ref)
	}
}

// sweep removes all unmarked objects from the heap.
func (h *Heap) sweep() []*GcObj {
	var deletedObjects []*GcObj
	var newObjects []*GcObj
	for _, obj := range h.objects {
		if obj.marked {
			// Keep the object and unmark it for the next GC cycle
			obj.marked = false
			newObjects = append(newObjects, obj)
		} else {
			fmt.Printf("Sweeping object: %s\n", obj.tv)
			deletedObjects = append(deletedObjects, obj)
		}
	}
	h.objects = newObjects
	return deletedObjects
}
