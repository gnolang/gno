// PKGPATH: gno.land/r/zrealm_fvshare_handler
// R2a: a writer shown or handed to another realm's handler comes back read-only.
package zrealm_fvshare_handler

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
	try("ProcessCtx", func() { fvshare.ProcessCtx(cross(cur), &R{}) })
	try("ProcessUndo", func() { fvshare.ProcessUndo(cross(cur), &R{}) })
	try("ProcessCapture", func() { fvshare.ProcessCapture(cross(cur), &R{}) })
	try("ProcessMethodValue", func() { fvshare.ProcessMethodValue(cross(cur), &R{}) })
	try("ProcessDefer", func() { fvshare.ProcessDefer(cross(cur), &R{}) })
	try("ProcessVariadic", func() { fvshare.ProcessVariadic(cross(cur), &R{}) })
	try("ProcessSpread", func() { fvshare.ProcessSpread(cross(cur), &R{}) })
}

// Output:
// ProcessCtx refused: true count: 0
// ProcessUndo refused: true count: 0
// ProcessCapture refused: true count: 0
// ProcessMethodValue refused: true count: 0
// ProcessDefer refused: true count: 0
// ProcessVariadic refused: true count: 0
// ProcessSpread refused: true count: 0
