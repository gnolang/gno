// Package launderpkg defines a struct type that the laundervictim
// realm uses as the type of its package-level state. Its methods write the
// receiver only when the calling code acts for the realm that owns it.
package launderpkg

type Object struct {
	Field string
}

// PInitData is a /p/-init-allocated package-level data var. Its
// StructValue carries ObjectInfo.PkgID = /p/demo/tests/launderpkg/v0.
// Used by zrealm_launder_pdata_* filetests to probe the /p/-source
// read patterns that the existing 62 /r/-source launder filetests
// don't cover. Empirically, direct value-read and pointer-deref of
// PInitData from a /r/-caller both panic readonly tainted — the
// /p/-source bytes are not adoptable into /r/-authority via these
// patterns.
var PInitData = Object{Field: "p-init"}

// Set is the canonical "/p/ helper that mutates /r/-owned state"
// pattern. Called by another realm's code it writes with that realm's
// authority, not the owner's, so a *Object handed out of a realm is a reader.
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

// Bare is a /p/-declared struct type with NO methods. Used to test
// whether embedding/fielding a methods-less /p/-type inside an
// /r/-declared container exposes any laundering vector through
// DIRECT field writes (no method dispatch, no Apply callback). The
// expectation: readonly taint on the /r/-container propagates to
// inner /p/-typed fields and direct writes panic.
type Bare struct {
	Field string
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
// caller-supplied callback. The signature looks read-only, and a callback
// from another realm gains no write authority through Apply. This is the
// avl.Tree.Iterate / list.ForEach shape.
func (i *Immutable) Apply(fn func(*Immutable)) {
	fn(i)
}

// BumpToPwn is a no-arg, no-return /p/-method that mutates the receiver,
// used by stored-bound-method-value probes: `mv := victimImmPtr.BumpToPwn`
// has type `func()`, which fits PlainHook = func(). Another realm's code
// invoking the bound value gets its write refused.
func (i *Immutable) BumpToPwn() {
	i.Field = "pwnd-via-bound-mv"
}

// DeferApply is the defer-variant of Apply: schedules fn(i) as a
// defer instead of calling synchronously. Used to probe whether
// the borrow rules apply the same when a callback is invoked from
// inside a /p/-function's defer queue.
func (i *Immutable) DeferApply(fn func(*Immutable)) {
	defer fn(i)
}

// PanicAfterApply: invokes fn(i) synchronously, then panics after
// return. If fn ran without panicking (write succeeded), the panic
// here is /p/-realm panic propagating up.
func (i *Immutable) PanicAfterApply(fn func(*Immutable)) {
	fn(i)
	panic("post-apply panic")
}

// RecoverApply: defers a recover(), then calls fn(i). If fn panics
// with readonly, recover catches it (within the same /p/-pkg
// frame). Returns the recovered value.
func (i *Immutable) RecoverApply(fn func(*Immutable)) (rec any) {
	defer func() {
		rec = recover()
	}()
	fn(i)
	return
}
