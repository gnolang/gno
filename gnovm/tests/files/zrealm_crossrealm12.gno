// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain"
	"chain/runtime"
	"fmt"

	psubtests "gno.land/p/demo/tests/subtests"
	rsubtests "gno.land/r/tests/vm/subtests"
)

func main(cur realm) {
	tests := []struct {
		fn func() runtime.Realm
	}{
		{unsafe.CurrentRealm},
		{psubtests.GetCurrentRealm},
		{func() runtime.Realm {
			return rsubtests.GetCurrentRealm(cross(cur))
		}},
	}

	for _, test := range tests {
		r := test.fn()

		if chain.PackageAddress(r.PkgPath()) != r.Address() {
			panic(fmt.Sprintf("ERROR: expected: %v, got: %v",
				chain.PackageAddress(r.PkgPath()), r.Address(),
			))
		}

		println(r.PkgPath(), r.Address())
	}
}

// Error:
// gno.land/r/crossrealm/zrealm_crossrealm12.gno:17:4-10: name unsafe not declared

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm12.gno:17:4: undefined: unsafe
