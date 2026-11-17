// PKGPATH: gno.land/r/launderrh_mv_stored_hook
package launderrh_mv_stored_hook

// Stored-hook laundering probe.
//
// 1. Attacker registers an /r/-attacker-declared closure into victim's
//    /r/-state via RegisterHook.
// 2. Attacker invokes victim's RunHooks(). borrow rule #1 → m.Realm = victim.
// 3. RunHooks loops over gHooks and invokes each. The stored closure
//    body is /r/-attacker-declared, so borrow rule #1 fires AGAIN: m.Realm =
//    /r/-attacker. Write tainted → readonly check fires → panic.
//
// Expected outcome: BLOCKED — the closure's borrow rule #1 shift back to the
// attacker realm is the safety net. Same trace as
// zrealm_launder_rdata_callback.gno, just delayed via registration.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	launderrvictim.ClearHooks(cross(cur))
	println("before:", launderrvictim.ReadImm())

	launderrvictim.RegisterHook(cross(cur), func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-stored-rh-closure"
	})
	launderrvictim.RunHooks(cross(cur))

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// gno.land/r/launderrh_mv_stored_hook's code was entered without crossing and is read-only (call its crossing function): i<VPBlock(1,0)>.Field
