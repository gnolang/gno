// PKGPATH: gno.land/r/zrealm_fvshare_nested
// A writer saved or shown before a nested crossing call never writes; nor
// does one never handed anywhere, as it is called as a value, not by name.
package zrealm_fvshare_nested

import "gno.land/r/tests/vm/fvshare"

var called int

func count(cur realm) { called++ }

func back(cur realm) { fvshare.Take(cross(cur), nil) }

func callSaved(cur realm) { fvshare.UseSaved(cross(cur)) }

func shownCall(cur realm, c *fvshare.Ctx) { c.Undo() }

func shownBack(cur realm, c *fvshare.Ctx) { fvshare.CallAny(cross(cur), c) }

func shownKeep(cur realm, c *fvshare.Ctx) {}

func try(name string, f func()) {
	r := revive(f)
	println(name, "refused:", r != nil, "count:", fvshare.Count(), "why:", r)
}

func main(cur realm) {
	fvshare.SaveClose(cross(cur))
	try("nested save", func() { fvshare.NestedSave(cross(cur), count) })
	try("nested callback", func() { fvshare.NestedCallBack(cross(cur), callSaved) })
	try("saved after return", func() { fvshare.UseSaved(cross(cur)) })
	try("shown, other realm calls", func() { fvshare.NestedShown(cross(cur), shownCall) })
	try("shown, called back", func() { fvshare.NestedShown(cross(cur), shownBack) })
	try("shown, run after nested", func() { fvshare.NestedShown(cross(cur), shownKeep) })
	try("private after nested", func() { fvshare.NestedPrivate(cross(cur), count) })
	try("private after nested back", func() { fvshare.NestedPrivate(cross(cur), back) })
	try("private", func() { fvshare.Private(cross(cur)) })
}

// Output:
// nested save refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// nested callback refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// saved after return refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// shown, other realm calls refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// shown, called back refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// shown, run after nested refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// private after nested refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// private after nested back refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// private refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
