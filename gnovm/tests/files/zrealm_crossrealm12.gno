// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"fmt"
	"std"

	psubtests "gno.land/p/demo/tests/subtests"
	rsubtests "gno.land/r/tests/vm/subtests"
)

func main(cur realm) {
	tests := []struct {
		fn func() std.Realm
	}{
		{std.CurrentRealm},
		{psubtests.GetCurrentRealm},
		{func() std.Realm {
			return rsubtests.GetCurrentRealm(cross)
		}},
	}

	for _, test := range tests {
		r := test.fn()

		if std.DerivePkgAddr(r.PkgPath()) != r.Address() {
			panic(fmt.Sprintf("ERROR: expected: %v, got: %v",
				std.DerivePkgAddr(r.PkgPath()), r.Address(),
			))
		}

		println(r.PkgPath(), r.Address())
	}
}

// Output:
// gno.land/r/crossrealm g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// gno.land/r/crossrealm g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// gno.land/r/tests/vm/subtests g17up60pmemnyemehamlm0x4tau9x2yhguq887pa
