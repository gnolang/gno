// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	"gno.land/r/tests/tests/crossrealm"
)

func Public(cur realm) {
	println("current realm: ", std.CurrentRealm())
	println("previous realm: ", std.PreviousRealm())
}

func main() {
	f := func(cur realm) {
		Public(cur)
	}

	crossrealm.SetClosure2(cross, f)

	crossrealm.ExecuteClosureCross(cross)
}

// Output:
// current realm:  CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
// previous realm:  CodeRealm{ g1wtde2f2gm29n4k3pdtzjqc4pqeyqj9upgwnsch, gno.land/r/tests/tests/crossrealm }
