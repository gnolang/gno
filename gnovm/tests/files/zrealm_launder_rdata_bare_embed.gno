// PKGPATH: gno.land/r/launderrh_be
package launderrh_be

// /r/-declared container EMBEDS a methods-less /p/-type
// (launderpkg.Bare). Attacker tries DIRECT field write through
// embedded-field promotion. No method dispatch, no Apply callback,
// no /p/-helper invocation. Pure attacker code mutating a field.
//
// Hypothesis: readonly taint on the /r/-container propagates to
// the inner /p/-typed field; the write check fires.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadBareEmbed())

	w := launderrvictim.GetWithBareEmbed()
	w.Field = "pwnd-direct-embed"

	println("after: ", launderrvictim.ReadBareEmbed())
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ package's object cannot be written after its init): w<VPBlock(1,1)>.Bare.Field
