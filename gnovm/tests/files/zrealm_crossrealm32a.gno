// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
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
// current realm:  CodeRealm{ g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p, gno.land/r/crossrealm_test }
// previous realm:  CodeRealm{ g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz, gno.land/r/demo/tests/crossrealm }
