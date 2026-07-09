// PKGPATH: gno.land/r/laundera
package laundera

// Attack A: attacker receives g by VALUE. Writes to the local copy.
//
// Post-#5747: the value-copy deep-duplicates the struct into a fresh,
// independent heap value owned by the executing realm (laundera). The
// local write succeeds and has no victim impact (pass-by-value
// semantics). The cross-realm-write defense is enforced at the
// pointer-deref boundary, not at the Copy boundary — see
// zrealm_launder_rdata_conv_val_then_addr for the *p = newval case
// that *is* blocked.

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
