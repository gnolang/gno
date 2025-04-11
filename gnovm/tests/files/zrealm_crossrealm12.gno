// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"chain"
	"chain/runtime"
	"fmt"

	psubtests "gno.land/p/demo/tests/subtests"
	rsubtests "gno.land/r/demo/tests/subtests"
)

func main() {
	tests := []struct {
		fn func() runtime.Realm
	}{
		{runtime.CurrentRealm},
		{psubtests.GetCurrentRealm},
		{rsubtests.GetCurrentRealm},
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
// gno.land/r/crossrealm_test g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// gno.land/r/crossrealm_test g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// gno.land/r/demo/tests/subtests g13g48xnr7lzxsrvny0uf6lhx0cfaxy4n0n5geuf
