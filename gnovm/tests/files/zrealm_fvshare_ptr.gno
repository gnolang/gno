// PKGPATH: gno.land/r/zrealm_fvshare_ptr
// R2a: a writer reached through a pointer a caller's handler chose is read-only.
package zrealm_fvshare_ptr

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
	try("SettlePtr", func() { fvshare.SettlePtr(cross(cur), &R{}) })
	try("SettleArg", func() { fvshare.SettleArg(cross(cur), &fvshare.HooksPtr().OnDone) })
	try("SettleLocal", func() { fvshare.SettleLocal(cross(cur), &R{}) })
}

// Output:
// SettlePtr refused: true count: 0
// SettleArg refused: true count: 0
// SettleLocal refused: true count: 0
