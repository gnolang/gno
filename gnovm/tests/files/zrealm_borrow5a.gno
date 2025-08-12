// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/demo/tests/crossrealm"
)

func main() {
	var ff = func() {
		crossrealm.AllowedList = append(crossrealm.AllowedList, 3)
		println(crossrealm.AllowedList)
	}

	crossrealm.S1.SetFn(ff)
	// cx expr is: s.F()
	crossrealm.S1.Execute()
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/r/demo/tests/crossrealm) package{})).AllowedList
