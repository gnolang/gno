// PKGPATH: gno.land/r/zrealm_import_mv
// A method value of a type outside the realm's import tree is view-only, even
// when the realm in whose code it is called binds it itself.
package zrealm_import_mv

import (
	"bytes"

	"gno.land/r/tests/vm/impvictim"
)

func main(cur realm) {
	impvictim.FillMV(cross(cur), bytes.NewReader([]byte("abcd")))
	println("victim:", impvictim.State())
}

// Error:
// cannot copy to readonly tainted slice (call a crossing function of the owning realm)
