// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/tests/vm/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm_test -> crossrealm
func main(cur realm) {
	f := func(_ int, rlm realm) {
		func(cur realm) {
			crossrealm.PrintRealms(cross(cur))
		}(cross(rlm))
	}

	crossrealm.SetClosure3(cross(cur), f)

	crossrealm.ExecuteClosure3(cross(cur))
}

// Output:
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
