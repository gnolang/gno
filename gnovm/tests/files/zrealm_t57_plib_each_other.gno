// PKGPATH: gno.land/r/tests/vm/t57eacho
package t57eacho

import other "gno.land/r/tests/vm/t57other"

var count int

// A98 attack: another realm's code hands the literal on to a library-owned
// receiver's method; whatever A98 decides, this call must stay refused.
func main(cur realm) {
	other.EachVia(func(int) { count++ })
	println(count)
}

// Error:
// gno.land/r/tests/vm/t57eacho's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/t57eacho) package{})).count
