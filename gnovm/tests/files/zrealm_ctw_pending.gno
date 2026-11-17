// PKGPATH: gno.land/r/ctw_pending
// An object the realm kept but has not saved yet is stored state: read-only
// code may not change it.
package ctw_pending

import "gno.land/r/tests/vm/ctw"



func main(cur realm) {
	ctw.StartX(cross(cur))
	ctw.IncPending()
}

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): l<VPBlock(1,0)>.N
