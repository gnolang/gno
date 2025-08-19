// PKGPATH: gno.land/r/borrow
package borrow

import (
	p_crossrealm "gno.land/p/demo/tests/p_crossrealm"
)

var f func()

func init() {
	f = p_crossrealm.FF
	f()
	println(p_crossrealm.AllowedList)
	println("========================================")
}

func main() {
	println(p_crossrealm.AllowedList)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/p/demo/tests/p_crossrealm) package{})).AllowedList
