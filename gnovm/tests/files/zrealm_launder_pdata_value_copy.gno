// PKGPATH: gno.land/r/launderpdata_v
package launderpdata_v

// /p/-source coverage: direct value-read of a /p/-init-allocated
// package global into a /r/-realm local.
//
// CLOSE_LAUNDER.md scenario 1 predicted this would launder the /p/-
// bytes to the caller's realm authority. The existing 62 launder
// filetests only exercised /r/-source paths; this fills the /p/-source
// gap.
//
// Empirically the attack does NOT succeed even with the CLOSE_LAUNDER
// preserve-/p/-stamp variant reverted. The readonly check fires at
// the LHS-pointer evaluation step (Machine.IsReadonly's RefValue
// branch or StructValue PkgID check, depending on path) before any
// Copy gets the chance to re-stamp. So the launder vector described
// in CLOSE_LAUNDER.md is closed by a layer below the Copy gate; the
// gate's /p/-extension was redundant for this scenario.
//
// This test pins the actual behavior: direct read-by-value of a /p/-
// global cannot be mutated by a /r/-caller, even post-CLOSE_LAUNDER-
// revert.

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
