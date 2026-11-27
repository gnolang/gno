// PKGPATH: gno.land/r/tests/vm/lendplibeach
package lendplibeach

import lib "gno.land/p/demo/tests/lendanylib"

var count int

// A literal lent to a library-owned receiver's method. The method runs
// as the library, so the literal is entered from another realm: refused.
func main(cur realm) {
	lib.Default.Each(func(int) { count++ })
	println(count)
}

// Error:
// gno.land/r/tests/vm/lendplibeach's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendplibeach) package{})).count
