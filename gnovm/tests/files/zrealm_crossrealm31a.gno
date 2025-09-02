// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/tests/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm
func main() {
	f := func() {
		crossrealm.PrintRealms(cross)
	}

	crossrealm.SetClosure(cross, f)

	crossrealm.ExecuteClosure(cross)
}

// Output:
// current realm: CodeRealm{ g1wtde2f2gm29n4k3pdtzjqc4pqeyqj9upgwnsch, gno.land/r/tests/crossrealm }
// previous realm: CodeRealm{ g1wtde2f2gm29n4k3pdtzjqc4pqeyqj9upgwnsch, gno.land/r/tests/crossrealm }
