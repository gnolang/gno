// PKGPATH: gno.land/r/ctw_owncb
// A literal in place writes only while a declared function reached by name (or
// via a library's own variable or table) runs it; a literal callee kept in a
// variable, or a declared one reached through a local, is refused.
package ctw_owncb

import (
	"gno.land/p/demo/tests/libtable/v0"
	"gno.land/r/tests/vm/ctw"
)

var count int

var boxes = []*libtable.Box{{N: 1}}

func main(cur realm) {
	println("apply:", revive(func() { ctw.ViaApply(cross(cur)) }), ctw.Count)
	libtable.Run(func() { count++ })
	println("apply twin by name:", count)
	ctw.ViaTable(cross(cur))
	println("table:", ctw.Count)
	println("local:", revive(func() { ctw.ViaLocal(cross(cur)) }), ctw.BoxSum())
	libtable.Each(boxes, func(b *libtable.Box) { b.N++ })
	println("local twin by name:", boxes[0].N)
	ctw.ViaDispatch(cross(cur))
	println("dispatch:", ctw.BoxSum())
}

// Output:
// apply: gno.land/r/tests/vm/ctw's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/ctw) package{})).Count 0
// apply twin by name: 1
// table: 1
// local: gno.land/r/tests/vm/ctw's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): b<VPBlock(1,0)>.N 3
// local twin by name: 2
// dispatch: 4
