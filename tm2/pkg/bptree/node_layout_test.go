package bptree

import (
	"reflect"
	"testing"
	"unsafe"
)

// ptrBytes returns the runtime's PtrBytes for a type: the offset past its last
// pointer word. The garbage collector scans exactly this many bytes of every
// object of the type, so for the big node structs it is the quantity that
// decides how much of a block's COW garbage the GC has to walk.
//
// reflect does not expose PtrBytes, so it is derived the same way the runtime
// computes it: the end of the last pointer-containing field, rounded up to a
// word.
func ptrBytes(t reflect.Type) uintptr {
	const word = unsafe.Sizeof(uintptr(0))
	var last uintptr
	var walk func(t reflect.Type, base uintptr)
	walk = func(t reflect.Type, base uintptr) {
		switch t.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Chan, reflect.Func:
			last = max(last, base+t.Size())
		case reflect.String, reflect.Slice:
			last = max(last, base+word) // data pointer is the first word
		case reflect.Interface:
			last = max(last, base+2*word) // itab + data
		case reflect.Struct:
			for i := range t.NumField() {
				f := t.Field(i)
				walk(f.Type, base+f.Offset)
			}
		case reflect.Array:
			// Every element has the same layout, so only the last one can
			// extend the pointer range.
			if t.Len() == 0 {
				return
			}
			elem := t.Elem()
			walk(elem, base+uintptr(t.Len()-1)*elem.Size())
		default:
		}
	}
	walk(t, 0)
	// Round up to a word, as the runtime does.
	if r := last % word; r != 0 {
		last += word - r
	}
	return last
}

// TestNodePointerBytes pins the pointers-first field grouping of InnerNode and
// LeafNode. Moving a pointer-bearing field below the hash/mini-merkle bulk
// would silently re-expand the GC's scan range over every COW clone — a pure
// throughput regression with no test of its own. The thresholds are generous
// (half the struct); the point is to catch a pointer field landing at the end,
// not to freeze exact offsets.
func TestNodePointerBytes(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  reflect.Type
	}{
		{"InnerNode", reflect.TypeOf(InnerNode{})},
		{"LeafNode", reflect.TypeOf(LeafNode{})},
	} {
		size := tc.typ.Size()
		pb := ptrBytes(tc.typ)
		t.Logf("%s: size=%d ptrBytes=%d (%.0f%%)", tc.name, size, pb, 100*float64(pb)/float64(size))
		if pb*2 > size {
			t.Errorf("%s: GC scans %d of %d bytes; group pointer fields first "+
				"(a pointer field below childHashes/valueHashes/miniTree re-expands the scan)",
				tc.name, pb, size)
		}
	}
}
