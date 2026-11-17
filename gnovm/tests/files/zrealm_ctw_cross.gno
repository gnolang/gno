// PKGPATH: gno.land/r/ctw_cross
// A crossing call writes, through the realm's helpers and methods too.
package ctw_cross

import "gno.land/r/tests/vm/ctw"



func main(cur realm) {
	ctw.Add(cross(cur), 2)
	println(ctw.Count, ctw.Saved.N)
}

// Output:
// 3 2
