// PKGPATH: gno.land/r/launderb
package launderb

// Attack B: attacker receives *g and tries to write directly through
// the pointer. Closed by the standard readonly check: base.PkgID =
// laundervictim, m.Realm = launderB (attacker), differ → readonly
// panic.

import (
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/laundervictim"
)

type ExploiterB struct{}

func (ExploiterB) Something(g launderpkg.Object) {}
func (ExploiterB) SomethingPtr(g *launderpkg.Object) {
	g.Field = "pwnd-via-ptr"
}

func main(cur realm) {
	laundervictim.Invoke(ExploiterB{})
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): g<VPBlock(1,1)>.Field
