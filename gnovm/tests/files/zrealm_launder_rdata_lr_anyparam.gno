// PKGPATH: gno.land/r/launderrh2
package launderrh2

// Attack L variant against /r/-DECLARED victim data, with an
// /r/-declared attacker impl.
//
// Unlike the /p/-attacker variant (zrealm_launder_rdata_l_anyparam.gno),
// the attacker here is in /r/ and CAN import /r/launderrvictim, so the
// type assertion `x.(*launderrvictim.Immutable)` succeeds. The question:
// does Rule 1 still block the write?
//
// Trace:
//   1. /r/launderrh2.main() — m.Realm = /r/launderrh2
//   2. launderrvictim.UseAnyMutator(...) — /r/-declared, Rule 1 →
//      m.Realm = /r/launderrvictim
//   3. UseAnyMutator body: m.Run(gImm) interface-dispatches to
//      EvilMutator.Run
//   4. EvilMutator.Run is /r/launderrh2-declared, Rule 1 fires →
//      m.Realm = /r/launderrh2
//   5. Run body: type-assert succeeds (same import access)
//   6. `t.Field = "pwnd"` writes /r/launderrvictim-stamped data
//      while m.Realm = /r/launderrh2 → readonly check fires
//
// Expected: panic at the write site, not at the assertion.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

type EvilMutator int

func (EvilMutator) Run(x any) {
	t := x.(*launderrvictim.Immutable)
	t.Field = "pwnd-from-rattacker"
}

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.UseAnyMutator(EvilMutator(0))

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): t<VPBlock(1,2)>.Field
