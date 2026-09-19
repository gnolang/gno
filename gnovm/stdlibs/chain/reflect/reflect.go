package reflect

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// X_objectInfo reads everything the VM knows about the object behind v.
//
// The parameter is a gno.TypedValue, which genstd links straight through with
// no Go2Gno conversion: the Gno side declares interface{} and the binding hands
// the value over untouched. That is what lets this native look at the object a
// value refers to instead of at a converted copy of it, and it needs no change
// to genstd: a Go gno.TypedValue parameter already matches any Gno parameter
// type (see mapping.signaturesMatch).
//
// It is a pure read. Nothing here allocates, persists, or advances a realm
// clock, so calling it has no effect on the object's lifecycle and repeated
// calls agree.
func X_objectInfo(m *gno.Machine, v gno.TypedValue) (id, addr, pkgPath, typ string, stamped, hasIdentity bool) {
	oo := ownObject(m, v)
	if oo == nil {
		return "", "", "", "", false, false
	}

	oid := oo.GetObjectID()
	if !oid.IsFinalized() {
		// Allocated, so the realm half of the ID is set, but the owning realm
		// has not persisted the object yet and the clock half is still zero.
		// It is an object; it does not have an identity yet, and an address
		// cannot be derived from half an ID.
		return "", "", "", "", false, true
	}

	return oid.String(),
		gno.DeriveObjectCryptoAddr(oid).String(),
		creatingRealmPath(m, oid),
		declaredTypeName(v),
		true, true
}

// ownObject returns the VM object that tv names in its own right, or nil.
//
// Only a pointer to a standalone heap item qualifies. Deliberately narrow:
// widening this later is a compatible change, while taking an answer back after
// its address has appeared in a balance is not.
//
// Notably not handled, and nil for that reason:
//
//   - A pointer whose base is a struct, an array or a block. Those are pointers
//     into a container (&x.Field, &arr[i], a pointer to a package level var),
//     and resolving them to the container would have every sibling report the
//     container's address as its own.
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
	switch cv := tv.V.(type) {
	case gno.PointerValue:
		// GetBase resolves a RefValue base through the store; a nil base,
		// which is what a nil pointer carries, comes back as nil.
		hiv, ok := cv.GetBase(m.Store).(*gno.HeapItemValue)
		if !ok {
			return nil
		}
		return hiv

	case *gno.FuncValue:
		// A func value is a reference, like a pointer: two variables holding
		// one closure hold the same object, and copying one copies the
		// reference. So a closure has an identity of its own and every holder
		// agrees on it, which is what makes a stored callback separately
		// addressable from whatever holds it.
		return cv

	case *gno.BoundMethodValue:
		// Same, for a method value bound to a receiver.
		return cv
	}

	return nil
}

// creatingRealmPath resolves the path of the realm whose storage clock minted
// oid. The ID carries only the realm's hashed PkgID, so this is a store lookup.
//
// It is a cache hit whenever that realm is loaded, which is the normal case:
// oid is finalized here, so the object is persisted, so the realm that owns it
// exists and is loaded for the object to have been reached at all. On a miss
// the store falls back to loading the realm's package value, which is present
// for the same reason.
//
// Returns "" rather than failing when the realm cannot be resolved, so a caller
// that only wanted the address is never denied it by a lookup it did not ask
// for.
func creatingRealmPath(m *gno.Machine, oid gno.ObjectID) string {
	if m.Store == nil {
		return ""
	}

	rlm := m.Store.GetRealmByID(oid.PkgID)
	if rlm == nil {
		return ""
	}

	return rlm.Path
}

// declaredTypeName renders the value's declared type as "<pkgpath>.<Name>".
//
// The value is a pointer, so the name wanted is the pointee's: a *Token reports
// the realm's Token, not "*Token". A type that is not a declared type, such as
// a pointer to an anonymous struct, has no such name and reports "".
func declaredTypeName(tv gno.TypedValue) string {
	pt, ok := tv.T.(*gno.PointerType)
	if !ok {
		return ""
	}

	dt, ok := pt.Elt.(*gno.DeclaredType)
	if !ok {
		return ""
	}

	return dt.PkgPath + "." + string(dt.Name)
}
