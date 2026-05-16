// PKGPATH: gno.land/r/curchaintest
package curchaintest

import (
	"gno.land/r/tests/vm/curchain_a"
)

// 3-level chain: curchaintest.main(cur) -> curchain_a.A(cross) -> curchain_b.B(cross).
// In B, cur is B; cur.Previous() is A; cur.Previous().Previous() is main; the
// next .Previous() is nil. Origin walks to main.
func main(cur realm) {
	println("main cur.PkgPath:", cur.PkgPath())
	curchain_a.A(cross2(cur))
}

// Output:
// main cur.PkgPath: gno.land/r/curchaintest
// A cur.PkgPath: gno.land/r/tests/vm/curchain_a
// A prev.PkgPath: gno.land/r/curchaintest
// B cur.PkgPath: gno.land/r/tests/vm/curchain_b
// B Previous().PkgPath: gno.land/r/tests/vm/curchain_a
// B Previous().Previous().PkgPath: gno.land/r/curchaintest
// B Previous()^3 PkgPath empty (EOA origin): true
// B Previous()^4 panics past origin: true
// B parity cur==Current: true
// B parity prev==Previous: true
