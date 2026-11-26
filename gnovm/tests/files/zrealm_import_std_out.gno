// PKGPATH: gno.land/r/zrealm_import_std_out
// A method reached through an interface is trusted only when its package is in
// the realm's import tree, standard library included.
package zrealm_import_std_out

import (
	"bytes"

	"gno.land/r/tests/vm/impvictim"
	"gno.land/r/tests/vm/impvictim/control"
)

func main(cur realm) {
	control.Fill(cross(cur), bytes.NewReader([]byte("efgh")))
	println("control:", control.State())
	impvictim.Fill(cross(cur), bytes.NewReader([]byte("abcd")))
	println("victim:", impvictim.State())
}

// Output:
// control: efgh

// Error:
// cannot copy to readonly tainted slice
