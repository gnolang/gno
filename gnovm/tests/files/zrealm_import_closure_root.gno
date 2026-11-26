// PKGPATH: gno.land/r/zrealm_import_closure_root
// A crossing closure handed out by a realm keeps that realm's trust root: the
// caller's imports are not trusted for code writing the closure's realm.
package zrealm_import_closure_root

import (
	"bytes"

	"gno.land/r/tests/vm/impvictim"
)

func main(cur realm) {
	f := impvictim.Filler()
	f(cross(cur), bytes.NewReader([]byte("abcd")))
	println("victim:", impvictim.State())
}

// Error:
// cannot copy to readonly tainted slice (call a crossing function of the owning realm)
