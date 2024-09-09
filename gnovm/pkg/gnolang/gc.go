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

func Unwrap(tv TypedValue) TypedValue {
	tvv := &tv

	for {
		ptr, ok := tvv.V.(PointerValue)

		if !ok {
			return *tvv
		}

		tvv = ptr.TV
	}
}

func MakeHeapObj(tv TypedValue) *GcObj {
	switch tv.V.(type) {
	case *SliceValue, *StructValue, *ArrayValue, StringValue:
		return &GcObj{
			marked: true,
			tv:     tv,
		}
	default:
		return nil
	}
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

func (h *Heap) AddRef(obj *GcObj, ref *GcObj) {
	switch v := ref.tv.V.(type) {
	case *StructValue:
		obj.refs = append(obj.refs, ref)

		for _, field := range v.Fields {
			fobj := h.FindObjectByTV(Unwrap(field))

			if fobj != nil {
				h.AddRef(ref, fobj)
			}
		}
	case *SliceValue:
		obj.refs = append(obj.refs, ref)
		av := v.Base.(*ArrayValue)

		for _, value := range av.List {
			fobj := h.FindObjectByTV(Unwrap(value))

			if fobj != nil {
				h.AddRef(ref, fobj)
			}
		}
	case StringValue:
		obj.refs = append(obj.refs, ref)
	default:
		panic(fmt.Sprintf("Unhandled type %T", v))
	}
}

func (h *Heap) FindObjectByTV(tv TypedValue) *GcObj {
	for _, object := range h.objects {
		if object.tv == tv {
			return object
		}
	}
	return nil
}

func (h *Heap) RemoveRoot(tv TypedValue) {
	roots := make([]*GcObj, 0, len(h.roots))
	var deleted bool

	for _, r := range h.roots {
		if !deleted && len(r.refs) == 1 && r.refs[0].tv == Unwrap(tv) {
			deleted = true
			continue
		}
		roots = append(roots, r)
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
			deletedObjects = append(deletedObjects, obj)
		}
	}
	h.objects = newObjects
	return deletedObjects
}
