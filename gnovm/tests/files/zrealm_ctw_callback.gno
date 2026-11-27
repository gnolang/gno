// PKGPATH: gno.land/r/ctw_callback
// A plain callback the realm calls back runs read-only: it may not write its
// own realm's state.
package ctw_callback

import "gno.land/r/tests/vm/ctw"

var sum int

func main(cur realm) {
	ctw.ForEach(cross(cur), func(n int) { sum += n })
}

// Error:
// gno.land/r/ctw_callback's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/ctw_callback) package{})).sum
