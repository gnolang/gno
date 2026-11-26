// PKGPATH: gno.land/r/zrealm_import_mv_caller
// A method value bound by the caller keeps its maker: a reader's method value
// passed into a realm that does not import the reader's package is view-only.
package zrealm_import_mv_caller

import (
	"bytes"

	"gno.land/r/tests/vm/impvictim"
)

func main(cur realm) {
	f := bytes.NewReader([]byte("abcd")).ReadAt
	impvictim.FillFunc(cross(cur), f)
	println("victim:", impvictim.State())
}

// Error:
// cannot copy to readonly tainted slice (call a crossing function of the owning realm)
