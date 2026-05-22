// PKGPATH: gno.land/r/launderpdata_v
package launderpdata_v

// /p/-source coverage: direct value-read of a /p/-init-allocated
// package global into a /r/-realm local.
//
// The existing 62 zrealm_launder_* filetests only exercised /r/-source
// paths. This fills the /p/-source gap: a /r/-caller reads a /p/-
// stamped global by value and tries to mutate the local copy.
//
// Empirically the mutation is rejected: the readonly check fires at
// the LHS-pointer evaluation step (Machine.IsReadonly's RefValue
// branch or StructValue PkgID check, depending on path) before
// StructValue.Copy gets the chance to re-stamp. The /p/-source bytes
// are not adoptable into /r/-authority via direct read; the protection
// layer sits below the Copy gate.

import (
	"gno.land/p/demo/tests/launderpkg"
)

func main(cur realm) {
	stolen := launderpkg.PInitData
	stolen.Field = "pwnd-via-pdata-value-launder"
	println(stolen.Field)
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): stolen<VPBlock(1,1)>.Field
