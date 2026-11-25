// PKGPATH: gno.land/r/zrealm_fvshare_copy
// R2a: what is stored or handed out is a copy without write power; the local
// name keeps it, unless the local itself was shown.
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

func try(name string, f func()) {
	println(name, "refused:", revive(f) != nil, "count:", fvshare.Count())
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
}

// Output:
// StoreLocal var refused: false count: 1
// StoreLocal map refused: false count: 2
// StoreLocal append in place refused: false count: 3
// StoreLocal append grown refused: false count: 4
// StoreLocal copy refused: false count: 5
// HandLocal realm refused: false count: 6
// HandLocal lib refused: false count: 7
// HandStruct local refused: false count: 8
// StoreBox refused: true count: 8
// StoreAlias refused: true count: 8
// HandStruct field refused: true count: 8
// ShownVar refused: true count: 8
// ShownResult refused: true count: 8
// ShownResult deferred refused: true count: 8
// PickKeys refused: true count: 8
// HandKeys refused: true count: 8
// ProcessPlant refused: true count: 8
