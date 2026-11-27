// PKGPATH: gno.land/r/tests/vm/lendplibeachother
package lendplibeachother

import other "gno.land/r/tests/vm/lendanyother"

var count int

// Attack: another realm's code hands the literal on to a library-owned
// receiver's method; whatever the loan rule decides, this call must stay refused.
func main(cur realm) {
	other.EachVia(func(int) { count++ })
	println(count)
}

// Error:
// gno.land/r/tests/vm/lendplibeachother's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendplibeachother) package{})).count
