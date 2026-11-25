// PKGPATH: gno.land/r/zrealm_fvshare_private
// R2a control: a writer made and used inside one crossing call writes; passed
// through a function value, even the realm's own stored helper, it does not.
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

var _ = fvpass.L{}

func try(name string, f func()) {
	println(name, "refused:", revive(f) != nil, "count:", fvshare.Count())
}

func main(cur realm) {
	try("Private", func() { fvshare.Private(cross(cur)) })
	try("Decorate", func() { fvshare.Decorate(cross(cur)) })
}

// Output:
// Private refused: false count: 1
// Decorate refused: true count: 1
