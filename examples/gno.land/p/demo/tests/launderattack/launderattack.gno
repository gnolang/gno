// Package launderattack imports launderpkg and provides functions
// that attempt to mutate a /p/-typed value passed by pointer. This is
// the "/p/attack imports /p/orig" variation: the goal is for code
// declared in this /p/ package to mutate a /r/-victim's instance of
// launderpkg.Object.
package launderattack

import "gno.land/p/demo/tests/launderpkg"

// TamperDirect writes through the pointer in its own body. The body
// runs under whatever m.Realm the caller had at PushFrameCall — for a
// top-level /p/ function with no receiver, that's the caller's realm.
// If the caller is attacker-realm, m.Realm at the write is
// attacker-realm and the readonly check fires on a foreign-stamped
// base.
func TamperDirect(o *launderpkg.Object, s string) {
	o.Field = s
}

// TamperViaMethod dispatches through the launderpkg.Object's own /p/
// method. PushFrameCall for Set sees a receiver stamped with the
// victim's realm (the pointer aliases victim's persisted state),
// triggering borrow rule 2: m.Realm becomes victim for the duration
// of Set. The write succeeds — but this requires launderpkg.Object
// to expose a Set method to begin with.
func TamperViaMethod(o *launderpkg.Object, s string) {
	o.Set(s)
}

// Tamper has the SAME underlying struct layout as launderpkg.Immutable.
// /p/launderpkg deliberately gave Immutable no mutator method —
// callers were supposed to be unable to write its Field. /p/attack
// declares its own type with the same layout and adds a mutator. An
// attacker that holds a *launderpkg.Immutable can convert it to
// *Tamper and invoke Tamper.Set — the conversion is purely a
// type-tag change, the pointer still aliases victim's persisted
// memory. PushFrameCall for Tamper.Set then sees recv stamped with
// victim's realm (the underlying object is unchanged) and
// borrow-routes m.Realm to victim. The write succeeds.
type Tamper struct {
	Field string
}

// Set is the mutator that launderpkg.Immutable deliberately did NOT
// expose. /p/attack adds it via the parallel-type trick.
func (t *Tamper) Set(s string) {
	t.Field = s
}

// Convert is the attacker's helper that does the type punning so the
// caller doesn't have to write the conversion inline.
func Convert(p *launderpkg.Immutable) *Tamper {
	return (*Tamper)(p)
}

// EvilMutator implements launderpkg.Mutator with a PRIMITIVE
// underlying type. Underlying-type matters: a struct/array/etc.
// receiver has a *StructValue that gets PkgID-stamped at allocation,
// triggering PushFrameCall's receiver-borrow rule to shift m.Realm
// back to /p/launderattack. A primitive-underlying type has no
// *StructValue and no PkgID — `recv.GetFirstObject` returns nil, so
// the borrow rule's `if obj != nil { ... }` branch is skipped and
// m.Realm stays at whatever the caller had it set to.
//
// When Run is dispatched via interface from inside a /p/-method
// body that was receiver-borrowed to the victim, m.Realm at Run's
// entry is the victim's — and stays the victim's, because EvilMutator
// (an int underneath) carries no PkgID to borrow against. The write
// inside Run commits under victim authority.
type EvilMutator int

func (EvilMutator) Run(i *launderpkg.Immutable) {
	i.Field = "pwnd-via-iface"
}

// EvilNilRecv tests the nil-pointer-receiver variant. Calling a method
// on a nil *EvilNilRecv is legal in Gno when the body doesn't deref
// the receiver. recv = PointerValue{Base: nil} → GetBase returns nil
// → GetFirstObject returns nil. Same "no anchor" gap as the
// primitive-receiver case, reachable through *T receivers.
type EvilNilRecv struct {
	X int // unused
}

func (n *EvilNilRecv) Run(i *launderpkg.Immutable) {
	// Does NOT dereference n. Just writes through target.
	i.Field = "pwnd-via-nilrecv"
}

// EvilFunc / EvilSlice / EvilMap — additional nil-anchor shapes.
// nil-valued receivers of defined types whose underlying is a
// reference type (slice/map/func) also have GetFirstObject == nil.
// The Attack H/I fix should cover all of them via the
// recvDeclaredTypePkgPath helper.
type EvilFunc func()

func (EvilFunc) Run(i *launderpkg.Immutable) { i.Field = "pwnd-via-func" }

type EvilSlice []int

func (EvilSlice) Run(i *launderpkg.Immutable) { i.Field = "pwnd-via-slice" }

type EvilMap map[string]int

func (EvilMap) Run(i *launderpkg.Immutable) { i.Field = "pwnd-via-map" }

// EvilSliceMutator's Run takes a slice — NOT a pointer parameter, so
// the Attack H/I fix's `hasForeignPPtrParam` predicate skips it.
// Tests whether the anchor predicate needs to look INSIDE composite
// parameter types for foreign-/p/ pointers.
type EvilSliceMutator int

func (EvilSliceMutator) Run(s []*launderpkg.Immutable) {
	s[0].Field = "pwnd-via-slice-arg"
}

// EvilAnyMutator's Run takes `any` and type-asserts to *Immutable.
// The signature reveals NO foreign-/p/-pointer statically — the
// pointer is hidden inside the interface box. Probes whether the
// predicate needs to fire on interface-typed params too.
type EvilAnyMutator int

func (EvilAnyMutator) Run(x any) {
	t := x.(*launderpkg.Immutable)
	t.Field = "pwnd-via-any-arg"
}

// EvilWrite is a top-level /p/-declared function value matching
// `func(*launderpkg.Immutable)`. Top-level /p/ functions trigger
// neither borrow rule #1 (not /r/-declared) nor borrow rule #2 (no receiver), so
// when EvilWrite is invoked as a callback from inside a
// borrowed-to-victim /p/-method body (e.g. Immutable.Apply, or any
// avl.Tree.Iterate-style hook), it inherits the victim's m.Realm
// and the write commits under victim authority.
func EvilWrite(i *launderpkg.Immutable) {
	i.Field = "pwnd-via-apply"
}

// EvilObjectWrite is the same shape but for *Object — used by tests
// that exercise Object's mutator surface through Apply-style callbacks.
func EvilObjectWrite(o *launderpkg.Object) {
	o.Field = "pwnd-via-apply-obj"
}

// Tamperer is a stamped /p/attack value (constructed at init under
// /p/attack's realm context, stamped /p/attack, persisted Frozen).
// Methods on Tamperer get receiver-borrowed to /p/attack at call
// time — useful as a control to compare against attacks where the
// receiver is victim-stamped.
type Tamperer struct{}

// TamperMethod is a /p/attack-defined method. When the caller invokes
// Tamperer{}.TamperMethod(o, s), the receiver Tamperer{} is
// constructed in the caller's realm, so receiver-borrow lands at the
// caller's realm, not /p/attack — the receiver carries the caller's
// stamp, not /p/attack's.
func (Tamperer) TamperMethod(o *launderpkg.Object, s string) {
	o.Field = s
}
