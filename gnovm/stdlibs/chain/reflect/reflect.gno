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
// address, and the realm whose clock minted that identity.
//
// It is opaque, comparable, and safe as a map key. Two Objects are equal if and
// only if they name the same object, however the value was reached: p,
// (*Alias)(p) and a p held behind an interface all give one Object. That is
// why an Object carries no type. The type of the value handed to [Of] is a
// property of the path the caller took, not of the object, and letting it into
// the comparison would make one object count once per spelling. The zero
// Object names nothing.
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
	pkgPath string // the realm whose clock minted id
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
// Object rather than inventing several.
//
// Always check ok. The zero Object's [Object.Address] is the empty address,
// which is not an account anyone can reach, and a transfer aimed at it is a
// transfer aimed at nothing. Of will not panic on your behalf: a realm that
// wants to refuse rather than continue should say so itself, as
// gnovm/tests/files/zrealm_reflect_proposal_fund.gno does.
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
// A func value, a bound method value or a map, which are references in the
// same sense a pointer is: two variables holding one closure, or one map, hold
// the same object, so every holder agrees on its address and a stored callback
// or index is addressable separately from whatever holds it.
//
// Only objects a realm persisted qualify. An object belonging to a stdlib or a
// p/ package, such as the func value strings.ToUpper, reports false, and
// [HasIdentity] reports false too.
//
// Everything else reports false, on purpose. A pointer into a struct field or
// an array element resolves to the container, and lending a value its
// container's identity would make every sibling answer the same address. A
// slice resolves to its backing array, which every other view of it shares. A
// struct or array arriving by value may be the persisted object or a copy of it
// depending on how it got here, which is the question this package will not
// guess at. A scalar, a nil pointer, a nil map and a nil interface have no
// object at all.
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
//
// "Later" means the same value, kept somewhere the realm persists. It does not
// mean evaluating the same expression again: &T{}, a func literal, a map
// literal and a method value such as tok.Method each allocate a fresh object
// every time they are evaluated, so Of(tok.Method) asked twice asks about two
// objects, neither of which is ever stored. Store the value first, then ask
// about what you stored.
func Of(v any) (obj Object, ok bool) {
	id, addr, pkgPath, stamped, _ := objectInfo(v)
	if !stamped {
		return Object{}, false
	}
	return Object{id: id, addr: addr, pkgPath: pkgPath}, true
}

// HasIdentity reports whether v is the kind of value that gets an identity of
// its own, whether or not the VM has stamped it yet.
//
// It exists to tell [Of]'s two false cases apart: "not yet" from "never". See
// Of for which values qualify.
//
// It is a second native call, and the native charges the same whichever answer
// it gives, so call it only on the branch where Of returned false. On the
// common path Of succeeds and one call has already produced everything the VM
// knows, this bool included.
func HasIdentity(v any) bool {
	_, _, _, _, has := objectInfo(v)
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
//
// The zero Object's address is the empty address. That is the one case to guard
// against, and checking [Of]'s ok is how.
func (o Object) Address() address { return address(o.addr) }

// ID returns the VM's own spelling of the object's identity,
// "<realm-id>:<clock-tick>", which is the same string the VM uses for the
// object in storage. It is "" for the zero Object.
//
// Use it to match an object against a storage dump. For naming an object
// anywhere else, prefer [Object.Address]. Do not parse it: the two halves are
// the VM's business, and this package will not promise their shape.
func (o Object) ID() string { return o.id }

// PkgPath returns the path of the realm whose storage clock minted the
// object's identity, which is the realm that owns its persistence.
//
// That is usually, but not always, the realm whose code allocated it. An
// object allocated by a maketx run script, or inside a stdlib function, has no
// lasting realm of its own, and the VM stamps it to the realm that persists it
// (see Realm.assignNewObjectID). Read PkgPath as "whose object this is", not as
// "whose code made it".
//
// Nor is it where the object's type is declared: a realm that instantiates a
// type from a p/ package creates an object of that type under its own path.
func (o Object) PkgPath() string { return o.pkgPath }

// IsZero reports whether o names no object.
func (o Object) IsZero() bool { return o.id == "" }

// String returns a line meant for a human: the owning realm, the clock tick
// that distinguishes the object from every other object of that realm, and the
// address. The ticks of a realm are not consecutive; do not read one as a
// count.
//
//	gno.land/r/demo/foo#7 (g1w4ek2u3jta047h6lta047h6lta047h6l9ht8xq)
//
// It is for logs and rendered pages. Emit [Object.Address] into events and use
// it as a key; this is not stable output to parse.
func (o Object) String() string {
	if o.IsZero() {
		return "<no object>"
	}
	return o.pkgPath + "#" + o.seq() + " (" + o.addr + ")"
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
// The exported API above is a rearrangement of these five values.
func objectInfo(v any) (id, addr, pkgPath string, stamped, hasIdentity bool) // injected
