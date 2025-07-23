// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/demo/tests/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm_test -> crossrealm
func main() {
	f := func() {
		func(cur realm) {
			crossrealm.PrintRealms(cross)
		}(cross)
	}

	crossrealm.SetClosure(cross, f)

	crossrealm.ExecuteClosure(cross)
}

// Output:
// current realm: CodeRealm{ g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz, gno.land/r/demo/tests/crossrealm }
// previous realm: CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
