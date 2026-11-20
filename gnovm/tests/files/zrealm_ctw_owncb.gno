// PKGPATH: gno.land/r/ctw_owncb
// Inside its own crossing call a realm writes through its own callback when
// the library code is named: a library's package variable or table entry, or
// the library's own dispatch. A local copy of a library function made in the
// gate acts like the named function, and so does the callback beneath it.
package ctw_owncb

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	ctw.ViaApply(cross(cur))
	ctw.ViaTable(cross(cur))
	println("count:", ctw.Count)
	ctw.ViaLocal(cross(cur))
	ctw.ViaDispatch(cross(cur))
	println("boxes:", ctw.BoxSum())
}

// Output:
// count: 2
// boxes: 6
