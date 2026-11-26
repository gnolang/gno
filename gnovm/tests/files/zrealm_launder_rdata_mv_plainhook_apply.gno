// PKGPATH: gno.land/r/launderrh_mv_plainhook_apply
package launderrh_mv_plainhook_apply

// Refused: a /p/ method has no power over another realm's object (R1).

// Stored PlainHook (func()) that captures a /p/-typed pointer aliasing
// victim memory and calls its /p/ Apply later with a /p/ callback.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	launderrvictim.ClearHooks(cross(cur))
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured := &we.Immutable

	launderrvictim.RegisterPlainHook(cross(cur), func() {
		captured.Apply(launderattack.EvilWrite)
	})
	launderrvictim.RunPlainHooks(cross(cur))

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
