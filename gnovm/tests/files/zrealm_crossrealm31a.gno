// PKGPATH: gno.land/r/crossrealm
package crossrealm

import "gno.land/r/tests/vm/crossrealm"

// crossrealm_test -> crossrealm -> crossrealm
func main(cur realm) {
	f := func(_ int, rlm realm) {
		crossrealm.PrintRealms(cross2(rlm))
	}

	crossrealm.SetClosure3(cross2(cur), f)

	crossrealm.ExecuteClosure3(cross2(cur))
}

// Output:
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
