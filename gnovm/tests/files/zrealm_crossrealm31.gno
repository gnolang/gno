// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/tests/vm/crossrealm"

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
// current realm: CodeRealm{ g1wtde2f2gm29n4k3pdtzjqc4pqeyqj9upgwnsch, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan, gno.land/r/crossrealm }
