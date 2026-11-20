// PKGPATH: gno.land/r/ctw_conv
// A realm's code entered without crossing may convert its own stored values:
// a conversion is a read.
package ctw_conv

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	println(ctw.RowCount(), ctw.RecN())
}

// Output:
// 3 7
