// Package reflect exposes the one part of the Gno VM's object model that a
// realm cannot observe from the language itself: object identity.
//
// This is not Go's reflect package and is not a step towards it. There is no
// Type, no Value, no field or method enumeration, and no way to read or write
// a value you were not handed. Anything that would let a realm inspect or
// mutate a foreign value belongs to a different discussion; this package
// answers one question, "which object is this", and nothing else.
//
// # Why the VM has to answer it
//
// A realm that needs a unique name for something it created has, today, only
// names it chooses itself: a path, a symbol, a counter it keeps. Anything a
// realm chooses, a realm can reuse, so two independent things can end up
// sharing one name and every consumer downstream loses the ability to tell
// them apart (gnolang/gno#6026).
//
// The VM already maintains exactly the name that cannot be reused: the
// ObjectID it stamps on an object when it persists it, minted from the owning
// realm's storage clock. Handing that out is strictly less machinery than any
// realm-side scheme, and it is unforgeable in a way no realm-side scheme is.
package reflect

// ObjectID is the identity the VM assigns to an object when it persists it.
//
// It is opaque, comparable, and safe as a map key. Two ObjectIDs are equal if
// and only if they name the same object. The zero ObjectID names no object,
// and is the only ObjectID for which IsZero reports true.
//
// An ObjectID cannot be constructed: [ObjectIDOf] is the only thing that
// returns a non-zero one. That is the property the whole package exists for.
// A realm choosing part of its own object's identity is the bug being fixed,
// so ObjectID is deliberately not a named string type, which any realm could
// mint out of any bytes it liked and pass off as VM-issued.
//
// An ObjectID is stable for the life of the object it names, across
// transactions. It is not an address: it has no signing key, nothing on chain
// can credit it, and no account exists at it. It identifies; it does not
// receive.
type ObjectID struct {
	// The VM's canonical "<realm-id>:<clock-tick>" spelling, empty when zero.
	// Unexported so that only ObjectIDOf can produce one.
	id string
}

// String returns the VM's canonical spelling of the ID, which is the same
// string the VM uses for the object in storage, or "" for the zero ObjectID.
//
// Use it to put an ID in an event or a log line. Do not parse it: the two
// halves are the VM's business, and this package will not promise their shape.
func (oid ObjectID) String() string { return oid.id }

// IsZero reports whether oid names no object.
func (oid ObjectID) IsZero() bool { return oid.id == "" }

// ObjectIDOf returns the identity of the object v refers to.
//
// ok is false whenever there is no identity to report, and [HasIdentity] says
// which of the two reasons applies:
//
//	!ok && HasIdentity(v)   v is an object, created by a call that has not
//	                        returned across a realm boundary yet, so the VM has
//	                        not stamped it. Asking again after that boundary
//	                        returns its ID. See "The stamping window" below.
//	!ok && !HasIdentity(v)  v is not the kind of value that gets an identity of
//	                        its own, and never will be. Nothing to wait for.
//
// When ok is false the returned ObjectID is the zero value, never a distinct
// placeholder, so a caller that ignores ok collapses every such value onto one
// ID rather than inventing several. Check ok.
//
// ObjectIDOf reads; it never writes. No realm clock advances, no object is
// created or persisted as a side effect, and calling it twice on the same
// value returns the same answer.
//
// # What has an identity
//
// A pointer to a standalone object, which is what `new(T)` and `&T{...}`
// produce. That covers the case this package is for: a realm hands out a
// *Thing and wants the VM's name for that Thing.
//
// Everything else reports false today, on purpose. A pointer into a struct
// field or an array element resolves to the container, and lending a value its
// container's identity would make every sibling answer the same ID. A slice
// resolves to its backing array, which every other view of it shares. A
// scalar, a nil pointer and a nil interface have no object at all.
//
// This list is meant to grow, and to only ever grow. Teaching a kind of value
// to report its identity is a compatible change; taking one back is not,
// because by then that ID is in somebody's event log. So a value whose
// identity semantics under copying and aliasing have not been settled reports
// false rather than an answer that might have to be withdrawn.
//
// # The stamping window
//
// An object's ObjectID is not complete when the object is created. The VM
// stamps the realm half at allocation and the clock half only when the owning
// realm persists the object, which happens when a realm frame returns. So an
// object created by the running call has no ID until that return.
//
// ObjectIDOf reports that window as ok == false rather than as an empty ID,
// which is the difference that matters: the caller has to decide what to do
// about it at the point where it is still possible to do something, instead of
// writing a blank into an event that is immutable the moment it is emitted.
// Two things a realm can do: emit the ID later, from a call that runs after
// the boundary; or refuse, with `if !ok { panic(...) }`, and make the caller
// split the work across two calls.
func ObjectIDOf(v interface{}) (oid ObjectID, ok bool) {
	id, stamped, _ := objectID(v)
	if !stamped {
		return ObjectID{}, false
	}
	return ObjectID{id: id}, true
}

// HasIdentity reports whether v is the kind of value that gets an ObjectID of
// its own, whether or not the VM has stamped it yet.
//
// It exists to tell [ObjectIDOf]'s two false cases apart: "not yet" from
// "never". See ObjectIDOf for which values qualify.
func HasIdentity(v interface{}) bool {
	_, _, has := objectID(v)
	return has
}

// objectID is the native binding. It returns the canonical ID string, whether
// the object has been stamped, and whether it is the kind of value that gets
// stamped at all. The exported API above is a pure rearrangement of these
// three; keeping the native single means one gas entry and one place where the
// VM's object model is read.
func objectID(v interface{}) (id string, stamped bool, hasIdentity bool) // injected
