// PKGPATH: gno.land/r/ctw_render
// Read-only code may change the temporaries it creates itself.
package ctw_render

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	println(ctw.Render(""))
}

// Output:
// rows:xx
