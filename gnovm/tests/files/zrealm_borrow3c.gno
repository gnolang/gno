// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/demo/tests/crossrealm"

func main() {
	crossrealm.SetClosure(cross, clos)

	crossrealm.ExecuteClosure(cross)
}

func clos() {
	crossrealm.AllowedList = append(crossrealm.AllowedList, 3)
	println(crossrealm.AllowedList)
}

// Output:
// slice[(1 int),(2 int),(3 int)]
