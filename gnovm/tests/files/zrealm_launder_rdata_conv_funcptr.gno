// PKGPATH: gno.land/r/launderrh_conv_funcptr
package launderrh_conv_funcptr

// CONVERSION VECTOR 7: function-type conversions and the /r/-victim
// hook surface.
//
// Hypothesis A: function-type conversion of an /r/-attacker function
// to a compatible func type is allowed (no foreign source, both types
// are functions).
// Hypothesis B: Passing the attacker function to victim.ApplyHook
// is statically allowed (callback param is /r/-declared but attacker
// can supply such a callback from its own realm). When victim runs
// h(gImm), the callback runs with attacker's m.Realm (Rule 1 from
// the /r/-attacker function's PkgID), so the write to victim-owned
// gImm.Field hits the readonly check.
//
// This test uses Error: directive because the readonly panic
// propagates as a fatal panic out of victim's call frame.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

// Attacker-defined callback whose signature matches ApplyHook's
// parameter type.
func evilHook(i *launderrvictim.Immutable) {
	i.Field = "pwnd-via-r-hook"
}

func main(cur realm) {
	// Convert (no-op) then dispatch via victim's hook surface.
	f := (func(i *launderrvictim.Immutable))(evilHook)
	launderrvictim.ApplyHook(f)
	println("after:", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
