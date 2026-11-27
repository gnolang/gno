// PKGPATH: gno.land/r/zrealm_fvshare_copy
// fvshare's 16 method-value routes are refused (a method value never writes);
// ProcessPlant: R.Plant is this realm's code entered without crossing, so it is
// read-only. A literal writer here writes in place via a declared callee, and is
// refused stored or handed out. Columns: fvshare count, this realm's count.
package zrealm_fvshare_copy

import (
	"gno.land/p/demo/tests/fvpass"
	"gno.land/r/tests/vm/fvshare"
)

type R struct{}

func (*R) Echo(p *func()) *func()      { return p }
func (*R) Pick(c *fvshare.Ctx) func()  { return c.Undo }
func (*R) Callback(undo func()) func() { return undo }
func (*R) PickKeys(m map[*fvshare.Ctx]bool) map[*fvshare.Ctx]bool {
	out := map[*fvshare.Ctx]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}
func (*R) Plant(gb func() *fvshare.Ctx, gw func() func()) { gb().Cb = gw() }

var (
	count  int
	onSale func()
)

func run(f func())                           { f() }
func keepRun(f func())                       { onSale = f; onSale() }
func handRun(h fvshare.Callbacker, f func()) { h.Callback(f)() }

// LitRoute runs one literal writer: in place, stored, or handed out.
func LitRoute(cur realm, route string) {
	switch route {
	case "in place":
		run(func() { count++ })
	case "stored":
		keepRun(func() { count++ })
	case "handed out realm":
		handRun(&R{}, func() { count++ })
	case "handed out lib":
		handRun(fvpass.L{}, func() { count++ })
	}
}

func try(name string, f func()) {
	println(name+":", revive(f), fvshare.Count(), count)
}

func main(cur realm) {
	for _, route := range []string{"var", "map", "append in place", "append grown", "copy"} {
		try("StoreLocal "+route, func() { fvshare.StoreLocal(cross(cur), route) })
	}
	try("HandLocal realm", func() { fvshare.HandLocal(cross(cur), &R{}) })
	try("HandLocal lib", func() { fvshare.HandLocal(cross(cur), fvpass.L{}) })
	try("HandStruct local", func() { fvshare.HandStruct(cross(cur), &R{}, false) })
	try("StoreBox", func() { fvshare.StoreBox(cross(cur)) })
	try("StoreAlias", func() { fvshare.StoreAlias(cross(cur)) })
	try("HandStruct field", func() { fvshare.HandStruct(cross(cur), &R{}, true) })
	try("ShownVar", func() { fvshare.ShownVar(cross(cur), &R{}) })
	try("ShownResult", func() { fvshare.ShownResult(cross(cur), &R{}, false) })
	try("ShownResult deferred", func() { fvshare.ShownResult(cross(cur), &R{}, true) })
	try("PickKeys", func() { fvshare.PickKeys(cross(cur), &R{}) })
	try("HandKeys", func() { fvshare.HandKeys(cross(cur), &R{}) })
	try("ProcessPlant", func() { fvshare.ProcessPlant(cross(cur), &R{}) })
	for _, route := range []string{"in place", "stored", "handed out realm", "handed out lib"} {
		try("LitRoute "+route, func() { LitRoute(cross(cur), route) })
	}
}

// Output:
// StoreLocal var: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreLocal map: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreLocal append in place: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreLocal append grown: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreLocal copy: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// HandLocal realm: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// HandLocal lib: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// HandStruct local: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreBox: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// StoreAlias: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// HandStruct field: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// ShownVar: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// ShownResult: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// ShownResult deferred: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// PickKeys: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// HandKeys: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N 0 0
// ProcessPlant: gno.land/r/zrealm_fvshare_copy's code was entered without crossing and is read-only (call its crossing function): gb<VPBlock(1,1)>().Cb 0 0
// LitRoute in place: undefined 0 1
// LitRoute stored: gno.land/r/zrealm_fvshare_copy's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_fvshare_copy) package{})).count 0 1
// LitRoute handed out realm: gno.land/r/zrealm_fvshare_copy's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_fvshare_copy) package{})).count 0 1
// LitRoute handed out lib: gno.land/r/zrealm_fvshare_copy's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_fvshare_copy) package{})).count 0 1
