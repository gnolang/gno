// PKGPATH: gno.land/r/laundera
package laundera

// Attack A: attacker receives g by VALUE. Writes to the local copy.
//
// Even though pass-by-value semantics already make the write affect
// only the attacker's local copy (no victim impact), CLOSE_LAUNDER's
// PkgID-preservation on StructValue.Copy means the local copy retains
// laundervictim's stamp. The attacker's write under m.Realm=laundera
// then trips the readonly check — the attack panics before it can do
// anything.
//
// This is the "patterns to audit before flipping the rule" case from
// CLOSE_LAUNDER.md: reads a /r/-global by value and treats the
// result as own state. Under CLOSE_LAUNDER, you cannot mutate even
// your own local copy if its origin stamp is a different realm.

import (
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/laundervictim"
)

type ExploiterA struct{}

func (ExploiterA) Something(g launderpkg.Object) {
	g.Field = "pwnd-via-value"
}
func (ExploiterA) SomethingPtr(g *launderpkg.Object) {}

func main(cur realm) {
	laundervictim.Invoke(ExploiterA{})
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): g<VPBlock(1,1)>.Field
