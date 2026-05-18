// Package launderpkg defines a struct type that the laundervictim
// realm uses as the type of its package-level state. The Set method
// is the "innocent /p/-helper" that an attacker tries to weaponize
// via the receiver-borrow rule (PushFrameCall borrow rule 2):
// when the receiver is owned by /r/X, calling Set borrows m.Realm to
// /r/X, and the write inside Set runs with /r/X authority.
package launderpkg

type Object struct {
	Field string
}

// Set is the canonical "/p/ helper that mutates /r/-owned state"
// pattern. Looks benign — like list.Set, avl.Set, etc. — but if a
// foreign caller obtains a pointer to a victim realm's Object, they
// can invoke Set on it and the borrow rule grants them victim
// authority for the call. Exposing a *Object out of a realm is
// equivalent to consenting to mutation by any caller that holds the
// pointer.
func (o *Object) Set(s string) {
	o.Field = s
}

// Read is a read-only accessor.
func (o *Object) Read() string {
	return o.Field
}

// Mutator is an interface — callers can pass any implementation. A
// realistic pattern: /p/orig defines a hook interface and exposes a
// method that lets callers register an impl for some operation.
type Mutator interface {
	Run(*Immutable)
}

// UseMutator is a /p/-method on *Object that dispatches an interface
// method (Mutator.Run) with target as the argument. This is the
// realistic shape — a /p/-library invoking a user-supplied hook on
// /r/-owned data.
func (o *Object) UseMutator(target *Immutable, m Mutator) {
	m.Run(target)
}

// SliceMutator's Run takes a SLICE of pointers — exercises the
// "non-pointer parameter that still conveys writable foreign data"
// shape that Attack K explores.
type SliceMutator interface {
	Run([]*Immutable)
}

// UseSliceMutator dispatches a SliceMutator on a single-element
// slice carrying target.
func (o *Object) UseSliceMutator(target *Immutable, m SliceMutator) {
	m.Run([]*Immutable{target})
}

// AnyMutator's Run takes `any` — interface-typed parameter. The
// caller can box a *Immutable into the interface and an attacker
// impl can type-assert back to write. Tests whether the predicate
// needs to treat interface-typed params as potentially foreign.
type AnyMutator interface {
	Run(any)
}

// UseAnyMutator boxes target into `any` before dispatching.
func (o *Object) UseAnyMutator(target *Immutable, m AnyMutator) {
	m.Run(target)
}

// Immutable is a deliberately read-only /p/ type: same struct layout
// as Object, but no mutator method. A realm using Immutable as the
// type of an exposed field intends "read-only API." The launder game
// variation explores whether an attacker /p/ package can convert
// *Immutable to a type with a mutator method declared elsewhere.
type Immutable struct {
	Field string
}

// Read is a read-only accessor.
func (i *Immutable) Read() string {
	return i.Field
}

// Apply is a higher-order helper that hands the *Immutable to a
// caller-supplied callback. The signature looks read-only (no
// mutator method on *Immutable itself), but Apply hands out an
// addressable pointer to victim-owned memory while m.Realm is
// borrowed to the victim. A /p/-declared callback substituted by
// the caller therefore runs with victim authority. This is the
// avl.Tree.Iterate / list.ForEach shape — common in /p/ libraries.
func (i *Immutable) Apply(fn func(*Immutable)) {
	fn(i)
}
