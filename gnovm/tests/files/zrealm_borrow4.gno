// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/tests/vm/crossrealm"

func main() {
	crossrealm.SetClosure(cross, clos)

	crossrealm.ExecuteClosure(cross)
}

func clos() {
	crossrealm.AllowedList = append(crossrealm.AllowedList, 3)
	println(crossrealm.AllowedList)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/r/tests/vm/crossrealm) package{})).AllowedList
