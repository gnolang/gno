// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm"
)

func Public(cur realm) {
	println("current realm: ", unsafe.CurrentRealm())
	println("previous realm: ", unsafe.PreviousRealm())
}

func main(cur realm) {
	f := func(cur realm) {
		Public(cur)
	}

	crossrealm.SetClosure2(cross(cur), f)

	crossrealm.ExecuteClosureCross(cross(cur))
}

// Output:
// current realm:  CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
// previous realm:  CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
