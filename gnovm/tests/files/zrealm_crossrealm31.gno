// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import "gno.land/r/demo/tests/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm_test -> crossrealm
func main() {
	f := func() {
		func(cur realm) {
			crossrealm.FooUpdate(cross, "a")
		}(cross)
	}

	crossrealm.SetClosure(cross, f)

	crossrealm.ExecuteClosure(cross)
}

// Output:
// current realm:  CodeRealm{ g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz, gno.land/r/demo/tests/crossrealm }
// previous realm:  CodeRealm{ g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p, gno.land/r/crossrealm_test }
