// PKGPATH: gno.land/r/laundera
package laundera

// Attack A: attacker receives g by VALUE. Writes to the local copy.
//
// Even though pass-by-value semantics already make the write affect
// only the attacker's local copy (no victim impact), StructValue.Copy
// preserves the source's PkgID for /r/-stamped values (gate is
// IsRealmPkg) — so the local copy retains laundervictim's stamp. The
// attacker's write under m.Realm=laundera then trips the readonly
// check — the attack panics before it can do anything.
//
// You cannot mutate even your own local copy when its origin stamp
// is a different /r/-realm.

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
