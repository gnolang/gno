// PKGPATH: gno.land/r/ctw_bump
// An exported non-crossing writer called from another realm is read-only.
package ctw_bump

import "gno.land/r/tests/vm/ctw"



func main(cur realm) {
	ctw.Bump()
}

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): Count<~VPBlock(3,0)>
