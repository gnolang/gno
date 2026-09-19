// Package reflect exposes what the Gno VM knows about a value's identity, and
// turns it into an address.
//
// This is not Go's reflect package and is not a step towards it. There is no
// Type, no Value, no field or method enumeration, and no way to read or write a
// value you were not handed. Type introspection over foreign values is a
// capability question of its own. This package answers one question, "which
// object is this, and where does it live", and gives the answer an address.
//
// # Why the VM has to answer it
//
// A realm that needs a unique name for something it created has, otherwise,
// only names it chooses itself: a path, a symbol, a counter it keeps, an ID its
// caller passes in. Anything a realm chooses, a realm can choose twice, and two
// things sharing one name are indistinguishable to everything downstream
// (gnolang/gno#6026).
//
// The VM already maintains the name that cannot be reused: the ObjectID it
// stamps on an object when it persists it, minted from the owning realm's
// storage clock. [Object.Address] hashes that into a bech32 address the same
// way a package path becomes a realm address.
//
// # What an address buys
//
// An object address is an ordinary gno.land address, so an object can be named
// anywhere an address is expected. Today that already means it can hold GRC20
// balances, since a token's Transfer takes a plain address, and it can receive
// ugnot. It cannot yet spend: a banker's sender is pinned to the realm's own
// package address, so letting a realm move funds held by an object it owns is a
// separate change. Until then, treat an object address as an account that
// receives and does not release.
package reflect

// Addressable is anything that can be named by an address: a user, a realm, or
// an object. It exists so an API can accept all three without caring which.
//
// Satisfying this interface is not authority. A realm can implement Address()
// returning any address at all, so use it to decide where value goes, never to
// decide who may move value. Authority over an address comes from the VM.
type Addressable interface {
	Address() address
}

// Object is what the VM knows about one persisted value: its identity, its
// address, the realm that created it, and its type.
//
// It is opaque, comparable, and safe as a map key. Two Objects are equal if and
// only if they name the same object. The zero Object names nothing.
//
// An Object cannot be constructed: [Of] is the only thing that returns a
// non-zero one. That is the property this package exists for, so Object is
// deliberately not a named string type, which any realm could mint out of any
// bytes it liked and pass off as VM-issued.
//
// An Object is a snapshot taken when [Of] returned. Everything in it is
// immutable for the life of the object it describes, so the snapshot never goes
// stale, and holding one does not keep the described object alive.
type Object struct {
	id      string // "<realm-id>:<clock-tick>", the VM's own spelling
	addr    string // bech32, derived from id
	pkgPath string // the realm that created the object
	typ     string // "<pkgpath>.<Name>" of the object's declared type
}

// Of returns what the VM knows about the object v refers to.
//
// ok is false whenever there is nothing to report, and [HasIdentity] says which
// of the two reasons applies:
//
//	!ok && HasIdentity(v)   v is an object the running call created, which the
//	                        VM has not persisted yet, so it has no identity and
//	                        no address. Asking again after the call returns
//	                        across a realm boundary gives both. See "The
//	                        stamping window" below.
//	!ok && !HasIdentity(v)  v is not the kind of value that gets an identity of
//	                        its own, and never will be. Nothing to wait for.
//
// When ok is false the returned Object is the zero value, never a distinct
// placeholder, so a caller that ignores ok collapses every such value onto one
// Object rather than inventing several. Check ok.
//
// Of reads; it never writes. No realm clock advances, no object is created or
// persisted as a side effect, and calling it twice returns the same answer.
//
// # What has an identity
//
// A pointer to a standalone object, which is what new(T) and &T{...} produce.
// That covers the case this package is for: a realm hands out a *Thing and
// wants the VM's name for that Thing.
//
// Everything else reports false, on purpose. A pointer into a struct field or
// an array element resolves to the container, and lending a value its
// container's identity would make every sibling answer the same address. A
// slice resolves to its backing array, which every other view of it shares. A
// scalar, a nil pointer and a nil interface have no object at all.
//
// This list is meant to grow, and to only ever grow. Teaching a kind of value
// to report its identity is a compatible change; taking one back is not,
// because by then that address is in somebody's balance. So a value whose
// identity under copying and aliasing has not been settled reports false.
//
// # The stamping window
//
// An object's identity is not complete when the object is created. The VM
// stamps the realm half at allocation and the clock half only when the owning
// realm persists the object, which happens when a realm frame returns. So an
// object created by the running call has no identity, and therefore no address,
// until that return.
//
// Of reports that window as ok == false rather than as an empty address, which
// is the difference that matters: the caller decides what to do about it at the
// point where doing something is still possible, instead of writing a blank
// into an event that is immutable the moment it is emitted, or handing out an
// unusable address. Two things a realm can do: read the address later, from a
// call that runs after the boundary; or refuse, and make the caller split the
// work across two calls.
func Of(v interface{}) (obj Object, ok bool) {
	id, addr, pkgPath, typ, stamped, _ := objectInfo(v)
	if !stamped {
		return Object{}, false
	}
	return Object{id: id, addr: addr, pkgPath: pkgPath, typ: typ}, true
}

// HasIdentity reports whether v is the kind of value that gets an identity of
// its own, whether or not the VM has stamped it yet.
//
// It exists to tell [Of]'s two false cases apart: "not yet" from "never". See
// Of for which values qualify.
func HasIdentity(v interface{}) bool {
	_, _, _, _, _, has := objectInfo(v)
	return has
}

// Address returns the object's address: a gno.land address derived from its VM
// identity, which the realm that owns the object chooses no part of.
//
// It is deterministic, unique per object, and stable for as long as the object
// lives. Two objects never share one, and an object address never collides with
// a realm's package address, because the two are hashed from different
// preimages.
//
// The address has no signing key, so nothing can sign for it. Coins sent to it
// are held by the object and, today, cannot be moved. See the package comment.
func (o Object) Address() address { return address(o.addr) }

// ID returns the VM's own spelling of the object's identity,
// "<realm-id>:<clock-tick>", which is the same string the VM uses for the
// object in storage. It is "" for the zero Object.
//
// Use it to match an object against a storage dump. For naming an object
// anywhere else, prefer [Object.Address]. Do not parse it: the two halves are
// the VM's business, and this package will not promise their shape.
func (o Object) ID() string { return o.id }

// PkgPath returns the path of the realm that created the object, which is the
// realm whose storage clock minted its identity and which owns its persistence.
//
// This is not necessarily where the object's type is declared: a realm that
// instantiates a type from a p/ package creates an object of that type under
// its own path. Compare with [Object.Type] when that distinction matters.
func (o Object) PkgPath() string { return o.pkgPath }

// Type returns the object's declared type as "<pkgpath>.<Name>", for example
// "gno.land/p/demo/tokens.Token". It is "" for an object whose type is not a
// declared type, such as a pointer to an anonymous struct.
func (o Object) Type() string { return o.typ }

// IsZero reports whether o names no object.
func (o Object) IsZero() bool { return o.id == "" }

// String returns a line meant for a human: the type, the sequence number that
// distinguishes objects of that type within their realm, and the address.
//
//	gno.land/r/demo/foo.Token#7 (g1w4ek2u3jta047h6lta047h6lta047h6l9ht8xq)
//
// It is for logs and rendered pages. Emit [Object.Address] into events and use
// it as a key; this is not stable output to parse.
func (o Object) String() string {
	if o.IsZero() {
		return "<no object>"
	}
	name := o.typ
	if name == "" {
		name = "<untyped>"
	}
	return name + "#" + o.seq() + " (" + o.addr + ")"
}

// seq returns the clock-tick half of the ID, the part that differs between two
// objects of one realm.
func (o Object) seq() string {
	for i := len(o.id) - 1; i >= 0; i-- {
		if o.id[i] == ':' {
			return o.id[i+1:]
		}
	}
	return o.id
}

// objectInfo is the native binding. One call, because a caller that wants the
// address almost always wants to record where the object came from next to it,
// and the realm lookup that PkgPath needs is a cache hit in the common case.
// The exported API above is a rearrangement of these six values.
func objectInfo(v interface{}) (id, addr, pkgPath, typ string, stamped, hasIdentity bool) // injected
