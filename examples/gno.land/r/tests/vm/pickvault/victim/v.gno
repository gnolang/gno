package victim

import "gno.land/p/nt/avl/v0"

// A vault only its crossing code should bump.

type Vault struct{ N int }

var vault = &Vault{}

func (v *Vault) Bump() { v.N++ }

func Get() *Vault { return vault }

func Count() int { return vault.N }

// Picker is a caller-supplied hook that picks a callback to run.
type Picker interface{ Pick() func() }

// RunPicked runs the callback a caller's picker returns.
func RunPicked(cur realm, p Picker) {
	f := p.Pick()
	f()
}

// DeferPicked defers the callback a caller's picker returns.
func DeferPicked(cur realm, p Picker) {
	f := p.Pick()
	defer f()
}

// Wrapper wraps a writer the victim hands it; it should return a wrapper.
type Wrapper interface{ Wrap(bump func()) func() }

// RunWrapped hands its own writer to a caller's wrapper and runs the result.
func RunWrapped(cur realm, w Wrapper) {
	f := w.Wrap(vault.Bump)
	f()
}

// RunAll runs every callback in a caller-supplied list.
func RunAll(cur realm, fs []func()) {
	for _, f := range fs {
		f()
	}
}

// RunVariadic runs every callback passed variadically.
func RunVariadic(cur realm, fs ...func()) {
	for _, f := range fs {
		f()
	}
}

// RunAny runs a callback found in an interface.
func RunAny(cur realm, x any) {
	if f, ok := x.(func()); ok {
		f()
	}
}

// BumpFn is a declared non-crossing writer.
func BumpFn() { vault.N++ }

var tree = avl.NewTree()

func Tree() *avl.Tree { return tree }

func Size() int { return tree.Size() }

// SetPicker picks a setter to run.
type SetPicker interface{ PickSet() func(string, any) bool }

// RunSet runs the setter a caller's picker returns.
func RunSet(cur realm, p SetPicker) {
	f := p.PickSet()
	f("k", 1)
}

var kept func()

// Keep stores the callback a caller's picker returns.
func Keep(cur realm, p Picker) { kept = p.Pick() }

// RunKept runs the stored callback.
func RunKept(cur realm) { kept() }

// PtrWrapper wraps a pointer to a writer the victim hands it.
type PtrWrapper interface{ WrapPtr(p *func()) *func() }

// RunWrappedPtr hands a pointer to its own fresh writer out and runs what comes back.
func RunWrappedPtr(cur realm, w PtrWrapper) {
	f := vault.Bump
	p := w.WrapPtr(&f)
	(*p)()
}
