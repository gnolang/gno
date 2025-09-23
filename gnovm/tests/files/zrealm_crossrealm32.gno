// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm"
)

func Public(cur realm) {
	println("current realm: ", runtime.CurrentRealm())
	println("previous realm: ", runtime.PreviousRealm())
}

func main() {
	f := func(cur realm) {
		Public(cross)
	}

	crossrealm.SetClosure2(cross, f)

	crossrealm.ExecuteClosureCross(cross)
}

// Output:
// current realm:  CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
// previous realm:  CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
