package reflect

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// X_objectInfo reads everything the VM knows about the object behind v.
//
// The parameter is declared any, the same as on the Gno side, and genstd hands
// over the machine's own gno.TypedValue boxed in it, with no Go2Gno conversion
// (see mappingType.BoxedTypedValue). That is what lets this native look at the
// object a value refers to instead of at a converted copy of it.
//
// Declaring gno.TypedValue here instead would read better and does link under
// the interpreter, but it does not survive `gno tool transpile --gobuild`: the
// transpiler rewrites objectInfo(v) into a direct Go call X_objectInfo(nil, v)
// that passes the Gno argument as the Gno side declares it, so the Go
// parameter has to be a type the Gno one is assignable to. Every other native
// in the tree already satisfies that by having the two sides spell the same
// type; this is the first one whose Gno type is an interface.
//
// The assertion below therefore fails in exactly one place: the Go program the
// transpiler emits, which has no VM behind it. There is no object identity
// outside the VM, so the honest answer there is the one for a value that has
// none.
//
// It is a pure read. Nothing here allocates, persists, or advances a realm
// clock, so calling it has no effect on the object's lifecycle and repeated
// calls agree.
//
// There is deliberately no type among the results. The only type at hand is
// the one the caller reached the object through, and one object is reachable
// through many: (*Alias)(p) and p name the same heap item. Reporting that type
// would make two Objects for one object compare unequal. Nor does the heap
// item keep a type of its own that could stand in: an assignment through a
// converted pointer overwrites the stored value's T (TypedValue.Assign copies
// tv2 whole), so even that changes over the object's life.
func X_objectInfo(m *gno.Machine, v any) (id, addr, pkgPath string, stamped, hasIdentity bool) {
	tv, ok := v.(gno.TypedValue)
	if !ok {
		return "", "", "", false, false
	}

	oo := ownObject(m, tv)
	if oo == nil {
		return "", "", "", false, false
	}

	oid := oo.GetObjectID()
	if !oid.IsFinalized() {
		// Allocated, so the realm half of the ID is set, but the owning realm
		// has not persisted the object yet and the clock half is still zero.
		// It is an object; it does not have an identity yet, and an address
		// cannot be derived from half an ID.
		return "", "", "", false, true
	}

	if !oid.PkgID.IsRealmPkg() {
		// Persisted by a stdlib or a p/ package, not by a realm: a top-level
		// func of strings, say. GetRealmByID would resolve it anyway and
		// report "strings" as the creating realm, which is not a realm. Whether
		// immutable-package objects should be addressable has not been
		// settled, so report "never"; widening that later is compatible,
		// withdrawing an address already in a balance is not. IsRealmPkg also
		// rejects a zero PkgID, the other half of DeriveObjectCryptoAddr's
		// precondition.
		return "", "", "", false, false
	}

	return oid.String(),
		gno.DeriveObjectCryptoAddr(oid).String(),
		creatingRealmPath(m, oid),
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
//   - A struct or an array arriving by value. TypedValue.Copy clones both, so
//     whether such a value is the persisted object or a copy of it depends on
//     how it got here, which is exactly the question this package will not
//     guess at.
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
		// Same, for a method value bound to a receiver. Note that evaluating
		// tok.Method allocates a fresh one every time, so it is the stored
		// method value that gets an identity, never the expression.
		return cv

	case *gno.MapValue:
		// A map is a reference too: TypedValue.Copy shares the *MapValue
		// between holders rather than cloning it, so every variable holding
		// one map holds the same object.
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
	rlm := m.Store.GetRealmByID(oid.PkgID)
	if rlm == nil {
		return ""
	}

	return rlm.Path
}
