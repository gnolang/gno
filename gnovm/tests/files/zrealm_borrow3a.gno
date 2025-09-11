// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/tests/vm/crossrealm"
)

func main() {
	f := func(cur realm) {
		crossrealm.AllowedList = append(crossrealm.AllowedList, 3)
		println(crossrealm.AllowedList)
	}

	crossrealm.SetClosure2(cross, f)

	crossrealm.ExecuteClosureCross(cross)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/r/tests/vm/crossrealm) package{})).AllowedList
