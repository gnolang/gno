// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain"
	"chain/runtime"
	"fmt"

	psubtests "gno.land/p/demo/tests/subtests"
	rsubtests "gno.land/r/demo/tests/subtests"
)

func main(cur realm) {
	tests := []struct {
		fn func() runtime.Realm
	}{
		{runtime.CurrentRealm},
		{psubtests.GetCurrentRealm},
		{func() runtime.Realm {
			return rsubtests.GetCurrentRealm(cross)
		}},
	}

	for _, test := range tests {
		r := test.fn()

		if chain.DerivePkgAddr(r.PkgPath()) != r.Address() {
			panic(fmt.Sprintf("ERROR: expected: %v, got: %v",
				chain.DerivePkgAddr(r.PkgPath()), r.Address(),
			))
		}

		println(r.PkgPath(), r.Address())
	}
}

// Output:
// gno.land/r/crossrealm g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// gno.land/r/crossrealm g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// gno.land/r/demo/tests/subtests g13g48xnr7lzxsrvny0uf6lhx0cfaxy4n0n5geuf
