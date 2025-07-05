// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/demo/tests/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm
func main() {
	f := func() {
		crossrealm.PrintRealms(cross)
	}

	crossrealm.SetClosure(cross, f)

	crossrealm.ExecuteClosure(cross)
}

// Output:
// current realm: CodeRealm{ g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz, gno.land/r/demo/tests/crossrealm }
// previous realm: CodeRealm{ g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz, gno.land/r/demo/tests/crossrealm }
