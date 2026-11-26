// PKGPATH: gno.land/r/tests/vm/t57make
package t57make

import lib "gno.land/p/demo/tests/t57lib"

var n int

// A library-owned receiver's method calls a closure the library stored, by
// value, as the library (A97): reading and filling a fresh object still work.
func main(cur realm) {
	n = lib.Default.Make()
	println(n, lib.Lib.Fill(5))
}

// Output:
// 7 5
