// PKGPATH: gno.land/r/demo/crosscrosspkg
package crosscrosspkg

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

// cross(cur) drives a cross-realm call into another package. The
// new cur minted for the callee's frame has the callee's pkgPath
// (gno.land/r/tests/vm/crossrealm) and its prev points back to
// this realm (gno.land/r/demo/crosscrosspkg). Compare with bare
// cross — semantically identical when rlm IS the current cur.

func main(cur realm) {
	println("== via bare cross ==")
	crossrealm.PrintRealms(cross(cur))
	println("== via cross(cur) ==")
	crossrealm.PrintRealms(cross(cur))
}

// Output:
// == via bare cross ==
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1xy6nk9gg0kvw76p68350f4m28h3apzjqrr38qx, gno.land/r/demo/crosscrosspkg }
// == via cross(cur) ==
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1xy6nk9gg0kvw76p68350f4m28h3apzjqrr38qx, gno.land/r/demo/crosscrosspkg }
