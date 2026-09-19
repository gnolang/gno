package reflect

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// X_objectID reads the identity of the object behind v.
//
// The parameter is a gno.TypedValue, which genstd links straight through with
// no Go2Gno conversion: the Gno side declares interface{} and the binding
// hands the value over untouched. That is what lets this native look at the
// object a value refers to instead of at a converted copy of the value, and it
// needs no change to genstd: a Go gno.TypedValue parameter already matches
// any Gno parameter type (see mapping.signaturesMatch).
//
// It is a pure read. Nothing here allocates, persists, or advances a realm
// clock, so calling it has no effect on the object's lifecycle and repeated
// calls agree.
func X_objectID(m *gno.Machine, v gno.TypedValue) (id string, stamped bool, hasIdentity bool) {
	oo := ownObject(m, v)
	if oo == nil {
		return "", false, false
	}

	oid := oo.GetObjectID()
	if !oid.IsFinalized() {
		// Allocated, so the realm half of the ID is set, but the owning realm
		// has not persisted the object yet and the clock half is still zero.
		// It is an object; it does not have an ID yet.
		return "", false, true
	}

	return oid.String(), true, true
}

// ownObject returns the VM object that tv names in its own right, or nil.
//
// Only a pointer to a standalone heap item qualifies. Deliberately narrow:
// widening this later is a compatible change, while taking an answer back
// after it has been written into an event log is not.
//
// Notably not handled, and nil for that reason:
//
//   - A pointer whose base is a struct, an array or a block. Those are
//     pointers into a container (&x.Field, &arr[i], a pointer to a package
//     level var), and resolving them to the container would have every sibling
//     report the container's identity as its own.
//   - A slice, which resolves to its backing array and so is shared by every
//     other view of that array.
//   - A struct, map, array, func or bound method arriving by value. Whether
//     such a value is the persisted object or a copy of it depends on how it
//     got here, which is exactly the question this package will not guess at.
//   - Anything with no object behind it: a scalar, a nil pointer, a nil
//     interface.
//
// This never reaches gno.TypedValue.GetFirstObject, which panics outright on a
// package RefValue and on a bare heap item. Every shape is handled here or
// reported as nil.
func ownObject(m *gno.Machine, tv gno.TypedValue) gno.Object {
	pv, ok := tv.V.(gno.PointerValue)
	if !ok {
		return nil
	}

	// GetBase resolves a RefValue base through the store; a nil base, which is
	// what a nil pointer carries, comes back as nil.
	hiv, ok := pv.GetBase(m.Store).(*gno.HeapItemValue)
	if !ok {
		return nil
	}

	return hiv
}
