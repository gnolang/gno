// PKGPATH: gno.land/r/tests/vm/lendplibmake
package lendplibmake

import lib "gno.land/p/demo/tests/lendanylib"

var n int

// A library-owned receiver's method calls a closure the library stored, by
// value, as the library: reading and filling a fresh object still work.
func main(cur realm) {
	n = lib.Default.Make()
	println(n, lib.Lib.Fill(5))
}

// Output:
// 7 5
