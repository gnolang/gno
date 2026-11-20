// PKGPATH: gno.land/r/demo/entryregistry
package entryregistry

import "gno.land/r/tests/vm/entries"

// A crossing function written in the owner's source keeps the owner's
// authority when stored in a registry and called by a stranger: the counter
// runs from 0 to 5.

func main(cur realm) {
	println("start:", entries.Count())
	inc := entries.Entry(0)
	for i := 0; i < 5; i++ {
		inc(cross(cur))
	}
	println("after:", entries.Count())
}

// Output:
// start: 0
// after: 5
