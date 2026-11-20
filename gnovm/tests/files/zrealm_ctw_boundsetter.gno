// PKGPATH: gno.land/r/cbatk
// A caller binds the victim's exported non-crossing setter into a library
// closure and hands it to the victim's crossing Run. The closure runs
// view-only, and the victim's setter beneath it stays read-only: library
// code a caller picked never regains the victim's authority.
package cbatk

import (
	"gno.land/p/demo/tests/cbattack"
	"gno.land/r/tests/vm/cbvictim"
)

func main(cur realm) {
	cbvictim.Run(cross(cur), cbattack.Bind(cbvictim.SetOwner, "attacker"))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
