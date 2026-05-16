// PKGPATH: gno.land/r/demo/cross2crosspkg
package cross2crosspkg

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

// cross2(cur) drives a cross-realm call into another package. The
// new cur minted for the callee's frame has the callee's pkgPath
// (gno.land/r/tests/vm/crossrealm) and its prev points back to
// this realm (gno.land/r/demo/cross2crosspkg). Compare with bare
// cross — semantically identical when rlm IS the current cur.

func main(cur realm) {
	println("== via bare cross ==")
	crossrealm.PrintRealms(cross2(cur))
	println("== via cross2(cur) ==")
	crossrealm.PrintRealms(cross2(cur))
}

// Output:
// == via bare cross ==
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1eds9l5vgkeud7zyup77uke2xtcqxnnuq7qg752, gno.land/r/demo/cross2crosspkg }
// == via cross2(cur) ==
// current realm: CodeRealm{ g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk, gno.land/r/tests/vm/crossrealm }
// previous realm: CodeRealm{ g1eds9l5vgkeud7zyup77uke2xtcqxnnuq7qg752, gno.land/r/demo/cross2crosspkg }
