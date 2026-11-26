package t51lib

import "gno.land/p/demo/tests/t51lib/keep"

// Controls: each parameter is only called or passed on.
func Run(f func()) { f() }

func Each(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

// Tree is avl-like: Iterate wraps cb in a literal passed to a recursive walk.
type Tree struct{ Keys []string }

func (t *Tree) Iterate(cb func(k string) bool) bool {
	return t.walk(0, func(i int) bool { return cb(t.Keys[i]) })
}

func (t *Tree) walk(i int, f func(int) bool) bool {
	if i >= len(t.Keys) {
		return false
	}
	if f(i) {
		return true
	}
	return t.walk(i+1, f)
}

// Perms is WithPermission-like.
type Perms struct{ Allowed string }

func (p *Perms) WithPermission(user string, cb func()) {
	if user != p.Allowed {
		panic("unauthorized")
	}
	cb()
}

type Runner interface{ Run(f func()) }

type Direct struct{}

func (Direct) Run(f func()) { f() }

// Attacks.
func Store(h *keep.Holder, f func()) { h.F = f }

func Pass(h *keep.Holder, f func()) { keep.Keep(h, f) }

func Launder(h *keep.Holder, f func()) {
	keep.Keep(h, f)
	h.Fire()
}

func Wrap(h *keep.Holder, f func()) {
	keep.Keep(h, func() { f() })
	h.Fire()
}

func Ret(f func()) func() { return f }

func Any(f func()) any { return f }

func Throw(f func()) { panic(f) }

func Rebind(set func(func()), run func()) {
	set(run)
	run()
}

func Alias(f func()) {
	g := f
	g()
}

// RunV calls f only; rest shifts argument positions against parameters.
func RunV(f func(), rest ...func()) { f() }

// Fn's method M, called as the method expression Fn.M(recv, g), puts the
// receiver among the call's arguments.
type Fn func()

func (Fn) M(g func()) { g() }

// CbFn is a named callback type, like avl's IterCbFn.
type CbFn func()

func RunN(f CbFn) { f() }

// VKeep stores a variadic element; LendV hands its parameter there.
func VKeep(h *keep.Holder, fs ...func()) { h.F = fs[0] }

func LendV(f func()) *keep.Holder {
	h := &keep.Holder{}
	VKeep(h, f)
	h.Fire()
	return h
}

func KeepV(f func()) *keep.Holder {
	h := &keep.Holder{}
	VKeep(h, f)
	return h
}

// RunS passes f on through a spread call, like authz's DoByPrevious.
func RunS(f func(), rest ...func()) { RunV(f, rest...) }
