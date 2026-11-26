// PKGPATH: gno.land/r/t73/escape_ok_arg
// Control: passing `cur` down as an argument, and literals that take a
// realm or return a non-realm value, are unchanged.
package escape_ok_arg

import "gno.land/p/demo/tests/t73lib"

func main(cur realm) {
	println(t73lib.Describe("lib", cur))
	read := func(_ int, r realm) string { return r.PkgPath() }
	println(read(0, cur))
	addr := func() address { return cur.Address() }
	println(addr() == cur.Address())
}

// Output:
// lib: gno.land/r/t73/escape_ok_arg
// gno.land/r/t73/escape_ok_arg
// true
