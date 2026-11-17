// PKGPATH: gno.land/r/ctw_owncb
// Inside its own crossing call a realm writes through its own callback when
// the library code is named: a library's package variable or table entry, or
// the library's own dispatch. A local copy of a library function is a value:
// it runs view-only, and so does the realm's callback beneath it.
package ctw_owncb

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	ctw.ViaApply(cross(cur))
	ctw.ViaTable(cross(cur))
	println("count:", ctw.Count)
	println("local:", revive(func() { ctw.ViaLocal(cross(cur)) }))
	ctw.ViaDispatch(cross(cur))
	println("boxes:", ctw.BoxSum())
}

// Output:
// count: 2
// local: gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>.N
// boxes: 4
