// PKGPATH: gno.land/r/zrealm_fvshare_private
// Made and called in one crossing call: a method value, a literal kept in a
// local slice, or a literal through a stored helper is refused; a literal in
// place as a declared callee's argument writes. Legal form: func(cur realm) + cross.
package zrealm_fvshare_private

import (
	"gno.land/p/demo/tests/fvpass"
	"gno.land/r/tests/vm/fvshare"
)

type R struct{}

func (*R) PickPtr() *func()             { return &fvshare.HooksPtr().OnDone }
func (*R) Echo(p *func()) *func()       { return p }
func (*R) Pick(c *fvshare.Ctx) func()   { return c.Undo }
func (*R) Callback(undo func()) func()  { return undo }
func (*R) Use(get func() func()) func() { return get() }
func (*R) PickV(fs ...func()) func()    { return fs[len(fs)-1] }
func (*R) Fill(c *fvshare.Ctx)          { c.Cb = c.Undo }

var count int

var wrap = func(f func()) func() { return f }

func run(f func()) { f() }

// PrivateLit, DecorateLit, LitInPlace: Private and Decorate with a literal
// writer, and the same literal run in place by a declared callee.
func PrivateLit(cur realm) {
	fs := []func(){func() { count++ }}
	fs[0]()
}

func DecorateLit(cur realm) {
	cb := wrap(func() { count++ })
	cb()
}

func LitInPlace(cur realm) { run(func() { count++ }) }

var _ = fvpass.L{}

func try(name string, f func()) {
	println(name+":", revive(f), fvshare.Count(), count)
}

func main(cur realm) {
	try("Private", func() { fvshare.Private(cross(cur)) })
	try("Decorate", func() { fvshare.Decorate(cross(cur)) })
	try("PrivateLit", func() { PrivateLit(cross(cur)) })
	try("DecorateLit", func() { DecorateLit(cross(cur)) })
	try("LitInPlace", func() { LitInPlace(cross(cur)) })
}

// Output:
// Private: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// Decorate: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// PrivateLit: gno.land/r/zrealm_fvshare_private's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_fvshare_private) package{})).count 0 0
// DecorateLit: gno.land/r/zrealm_fvshare_private's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_fvshare_private) package{})).count 0 0
// LitInPlace: undefined 0 1
