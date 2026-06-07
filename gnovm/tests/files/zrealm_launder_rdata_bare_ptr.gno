// PKGPATH: gno.land/r/launderrh_bp
package launderrh_bp

// /r/-declared container with POINTER FIELD to methods-less /p/-type.
// Attacker tries direct field write through the pointer field.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadBarePtr())

	w := launderrvictim.GetWithBarePtr()
	w.Inner.Field = "pwnd-direct-ptr"

	println("after: ", launderrvictim.ReadBarePtr())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): w<VPBlock(1,1)>.Inner.Field
